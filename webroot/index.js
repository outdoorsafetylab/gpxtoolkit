function generate(format, suffix) {
    document.getElementById("format").value = format
    document.getElementById("filename-suffix").value = suffix
    document.getElementById("milestone-form").submit()
}