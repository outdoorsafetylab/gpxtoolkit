package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func NewParser() *Parser {
	return &Parser{Stack: Stack{slice: make([]string, 0)}, hooks: make(map[string]map[string]*xmlHook)}
}

type xmlEnterCallback func(map[string]string) error

type xmlTextCallback func(text string) error

type xmlLeaveCallback func() error

type xmlHook struct {
	enter xmlEnterCallback
	text  xmlTextCallback
	leave xmlLeaveCallback
}

type Parser struct {
	Stack
	hooks map[string]map[string]*xmlHook
	any   *xmlHook
}

func (s *Parser) OnAny(enter xmlEnterCallback, text xmlTextCallback, leave xmlLeaveCallback) *Parser {
	s.any = &xmlHook{enter: enter, text: text, leave: leave}
	return s
}

func (s *Parser) On(xpath string, enter xmlEnterCallback, text xmlTextCallback, leave xmlLeaveCallback) *Parser {
	// Strip the "//" prefix from xpath for consistent hook registration
	// This ensures hooks are stored without the prefix, matching our lookup logic
	xpath = strings.TrimPrefix(xpath, "//")

	splits := strings.Split(xpath, "/")
	last := splits[len(splits)-1]
	hooks := s.hooks[last]
	if hooks == nil {
		hooks = make(map[string]*xmlHook)
		s.hooks[last] = hooks
	}
	h := hooks[xpath]
	if h == nil {
		hooks[xpath] = &xmlHook{enter: enter, text: text, leave: leave}
	} else {
		if enter != nil {
			h.enter = enter
		}
		if text != nil {
			h.text = text
		}
		if leave != nil {
			h.leave = leave
		}
	}
	return s
}

func (s *Parser) OnEnter(xpath string, cb xmlEnterCallback) *Parser {
	s.On(xpath, cb, nil, nil)
	return s
}

func (s *Parser) OnLeave(xpath string, cb xmlLeaveCallback) *Parser {
	s.On(xpath, nil, nil, cb)
	return s
}

func (s *Parser) OnText(xpath string, strip bool, text xmlTextCallback) *Parser {
	if strip {
		s.On(xpath, nil, func(txt string) error {
			return text(strings.Trim(txt, " \t\n\r"))
		}, nil)
	} else {
		s.On(xpath, nil, text, nil)
	}
	return s
}

func (s *Parser) Parse(r io.Reader) error {
	started := false
	decoder := xml.NewDecoder(r)
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if t == nil {
			break
		}
		switch e := t.(type) {
		case xml.StartElement:
			attrs := make(map[string]string)
			for _, a := range e.Attr {
				attrs[a.Name.Local] = a.Value
			}
			err := s.push(e.Name.Local, attrs)
			if err != nil {
				return err
			}
			started = true
		case xml.EndElement:
			err := s.pop(e.Name.Local)
			if err != nil {
				return err
			}
		case xml.CharData:
			text := strings.Trim(string(e), " \r\n\t")
			err := s.text(text)
			if err != nil {
				return err
			}
		}
	}
	if !started {
		return fmt.Errorf("invalid XML")
	}
	return nil
}

func (s *Parser) push(e string, attrs map[string]string) error {
	s.Push(e)
	if s.any != nil && s.any.enter != nil {
		err := s.any.enter(attrs)
		if err != nil {
			return err
		}
	}
	hook := s.hook()
	if hook != nil && hook.enter != nil {
		return hook.enter(attrs)
	}
	return nil
}

func (s *Parser) text(text string) error {
	hook := s.hook()
	if hook != nil && hook.text != nil {
		return hook.text(text)
	}
	if s.any != nil && s.any.text != nil {
		err := s.any.text(text)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Parser) pop(e string) error {
	i := len(s.slice) - 1
	last := s.slice[i]
	if last != e {
		return fmt.Errorf("Pop mismatch: %s vs %s", last, e)
	}

	// Call leave hooks before popping the element so the stack state is correct
	hook := s.hook()
	if hook != nil && hook.leave != nil {
		err := hook.leave()
		if err != nil {
			return err
		}
	}

	if s.any != nil && s.any.leave != nil {
		err := s.any.leave()
		if err != nil {
			return err
		}
	}

	// Pop the element after calling hooks
	s.Pop()
	return nil
}

func (s *Parser) hook() *xmlHook {
	hooks := s.hooks[s.Peek()]
	if hooks != nil {
		// Strip the "//" prefix from XPath for hook lookup since hooks are registered without it
		xpath := strings.TrimPrefix(s.XPath(), "//")
		return hooks[xpath]
	}
	return nil
}

func (s *Parser) Dump() string {
	var b bytes.Buffer
	for _, e := range s.slice {
		b.WriteString("<")
		b.WriteString(e)
		b.WriteString(">")
	}
	return b.String()
}

func (s *Parser) XPath() string {
	var b bytes.Buffer
	for i, e := range s.slice {
		if i == 0 {
			b.WriteString("//")
		} else {
			b.WriteString("/")
		}
		b.WriteString(e)
	}
	return b.String()
}
