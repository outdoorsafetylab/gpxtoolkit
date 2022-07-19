<template>
  <div class="container-fluid min-vh-100 d-flex flex-column" @dragover.prevent @drop.prevent>
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark row">
      <div class="container-fluid">
        <div>
          <span v-if="gpxFile" class="navbar-text">{{ gpxFile.name }}</span>
          <button v-if="gpxFile" type="button" class="btn btn-danger" @click="clear()">清除</button>
          <input class="form-control" type="file" id="gpx-file" name="gpx-file" accept=".gpx" required
            @change="onGpxChosen" style="display:none;">
          <label v-if="!gpxFile" for="gpx-file" class="btn btn-primary position-relative">
            選擇GPX檔案...
            <span class="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">
              或拖放到地圖上
            </span>
          </label>
        </div>
        <div>
          <button v-if="gpxFile" type="button" class="btn btn-primary" @click="preview()">預覽</button>
          <div v-if="previewed" class="btn-group" role="group">
            <button id="btnGroupDrop1" type="button" class="btn btn-success dropdown-toggle" data-bs-toggle="dropdown"
              aria-expanded="false">
              下載
            </button>
            <div class="dropdown-menu">
              <button type="button" class="dropdown-item" @click="downloadGpx()">GPX</button>
              <button type="button" class="dropdown-item" @click="downloadCsv()">CSV</button>
            </div>
          </div>
          <button v-if="gpxFileContent && previewed" type="button" class="btn btn-danger" @click="restore()">還原</button>
          <div class="btn-group" role="group">
            <button id="btnGroupDrop1" type="button" class="btn btn-info dropdown-toggle" data-bs-toggle="dropdown"
              aria-expanded="false">
              關於
            </button>
            <div class="dropdown-menu dropdown-menu-end">
              <button type="button" class="dropdown-item" data-bs-toggle="modal"
                data-bs-target="#introModal">簡介及源起</button>
              <button type="button" class="dropdown-item" data-bs-toggle="modal"
                data-bs-target="#notesModal">注意事項</button>
            </div>
          </div>
        </div>
      </div>
    </nav>
    <div class="row progress" style="height: 4px;">
      <div v-if="progress > 0" class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar"
        v-bind:style="{ width: progress + '%' }"></div>
    </div>
    <div v-if="gpxFile" id="options" class="row pt-1 pb-2">
      <div class="col-md-2">
        <label for="distance" class="form-label">里程間距</label>
        <div class="input-group">
          <input type="number" class="form-control" id="distance" required aria-describedby="distance-addon2"
            v-model="distance">
          <span class="input-group-text" id="distance-addon2">公尺</span>
        </div>
      </div>
      <div class="col-md-5">
        <label for="template" class="form-label">里程航點名稱樣版</label>
        <div class="input-group">
          <input type="text" class="form-control" id="template" required aria-describedby="template-addon2"
            v-model="template">
          <button class="btn btn-outline-primary" type="button" id="template-addon2" data-bs-toggle="modal"
            data-bs-target="#templateModal">說明</button>
        </div>
      </div>
      <div class="col-md-2">
        <label for="symbol" class="form-label">里程航點符號 &lt;sym&gt;</label>
        <div class="input-group">
          <input type="text" class="form-control" id="symbol" required aria-describedby="symbol-addon2"
            v-model="symbol">
          <a class="btn btn-outline-primary" type="button" id="symbol-addon2"
            href="https://www.gpsrchive.com/BaseCamp/Custom%20Waypoint%20Symbols.html" target="_blank">說明</a>
        </div>
      </div>
      <div class="col-md-3">
        <div>
          <input type="checkbox" class="form-check-input" id="fits" v-model="fits">
          <label class="form-check-label" for="fits">遷就地標(路標架設用)</label>
        </div>
        <div>
          <input type="checkbox" class="form-check-input" id="reverse" v-model="reverse">
          <label class="form-check-label" for="reverse">反向產生里程航點</label>
        </div>
        <div>
          <input type="checkbox" class="form-check-input" id="terrainDistance" v-model="terrainDistance">
          <label class="form-check-label" for="terrainDistance">使用地表距離</label>
        </div>
      </div>
    </div>
    <div id="map" class="row flex-grow-1" @drop="onGpxDropped"></div>
    <div class="modal fade" id="introModal" tabindex="-1" aria-labelledby="introModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="introModalLabel">簡介及源起 (版本 {{ version }})</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="關閉"></button>
          </div>
          <div class="modal-body">
            <p>本工具可讀取原始GPX檔案，依指定間距計算出里程航點 (例如 0.1K, 0.2K 等等) 後產生新的GPX檔案。</p>
            <p>本工具源於桃園市山岳協會及中華民國山岳協會持續推動<a href="https://www.tytaaa.org.tw/news/7"
                target="_blank">登山路標的標準化及建置</a>，為簡化及加速路標架設的GPX前置處理作業，因此開始創作此工具。</p>
            <p>相關問題回報及建議請來信：<code>outdoorsafetylab 小老鼠 gmail.com</code></p>
            <p>如果您有 web 前端或 golang 後端技能，也歡迎在<a href="https://github.com/outdoorsafetylab/gpxtoolkit"
                target="_blank">GitHub</a>上加入我們持續改善此工具！</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" data-bs-dismiss="modal">關閉</button>
          </div>
        </div>
      </div>
    </div>
    <div class="modal fade" id="notesModal" tabindex="-1" aria-labelledby="notesModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="notesModalLabel">注意事項</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="關閉"></button>
          </div>
          <div class="modal-body">
            <ol>
              <li>請注意路線起點位置，里程會以其為基準開始計算。</li>
              <li>請預先去除雜點，漂移點及原地不動時的毛線球航跡。</li>
              <li>請務必再以人工確認產出結果。</li>
              <li>建議一條路線使用一個GPX檔案以便確認。</li>
            </ol>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" data-bs-dismiss="modal">關閉</button>
          </div>
        </div>
      </div>
    </div>
    <div class="modal fade" id="templateModal" tabindex="-1" aria-labelledby="templateModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="templateModalLabel">里程航點名稱樣版說明</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="關閉"></button>
          </div>
          <div class="modal-body">
            <p>樣版採用 <a href="https://go.dev/play/" target="_blank">Go 語言語法</a>，樣版可使用 <code>printf</code> 函數來調用
              <code>fmt.Printf</code>。並提供以下變數：
            </p>
            <ul>
              <li><b>num</b>: 里程號碼，整數，從1開始。</li>
              <li><b>total</b>: 里程航點總數，整數。</li>
              <li><b>dist</b>: 里程距離，浮點數(小數)，以公尺計算。</li>
              <li><b>lat</b>: 經度，浮點數(小數)，WGS84模型。</li>
              <li><b>lon</b>: 緯度，浮點數(小數)，WGS84模型。</li>
              <li><b>elev</b>: 標高，浮點數(小數)，以公尺計算。</li>
            </ul>
            <p>若需要轉換為英呎或英哩，可使用數學運算符號。例如 <code>dist*0.000621371192</code> 即將公尺轉換為英哩。</p>
            <p>範例：<code>dist</code><br />產出結果：<b>100</b>, <b>200</b>, <b>300</b>...</p>
            <p>範例：<code>dist/1000</code><br />產出結果：<b>0.1</b>, <b>0.2</b>, <b>0.3</b>...</p>
            <p>範例：<code>printf("%.0fm", dist)</code><br />產出結果：<b>100m</b>, <b>200m</b>, <b>300m</b>...</p>
            <p>範例：<code>printf("%.1fK", dist/1000)</code><br />產出結果：<b>0.1K</b>, <b>0.2K</b>, <b>0.3K</b>...</p>
            <p>範例：<code>printf("%.1fK/%.0fh", dist/1000, elev)</code><br />產出結果：<b>0.1K/2763h</b>, <b>0.2K/2756h</b>,
              <b>0.3K/2748h</b>...
            </p>
            <p>範例：<code>printf("%02d/%d", num, total)</code><br />產出結果：<b>01/86</b>, <b>02/86</b>, <b>03/86</b>...</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" data-bs-dismiss="modal">關閉</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap";
import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import MapboxLanguage from "@mapbox/mapbox-gl-language";
import toGeoJSON from "@mapbox/togeojson";
import { useCookies } from "vue3-cookies";

function encodeQueryData(data) {
  const ret = [];
  for (let d in data) {
    ret.push(encodeURIComponent(d) + '=' + encodeURIComponent(data[d]));
  }
  return ret.join('&');
}
function isDOMParseError(parsedDocument) {
  // parser and parsererrorNS could be cached on startup for efficiency
  let parser = new DOMParser(),
    errorneousParse = parser.parseFromString('<', 'application/xml'),
    parsererrorNS = errorneousParse.getElementsByTagName("parsererror")[0].namespaceURI;

  if (parsererrorNS === 'http://www.w3.org/1999/xhtml') {
    // In PhantomJS the parseerror element doesn't seem to have a special namespace, so we are just guessing here :(
    return parsedDocument.getElementsByTagName("parsererror").length > 0;
  }
  return parsedDocument.getElementsByTagNameNS(parsererrorNS, 'parsererror').length > 0;
}

mapboxgl.accessToken = "pk.eyJ1Ijoib3V0ZG9vcnNhZmV0eWxhYiIsImEiOiJjbDU1Y2N1eW0wbTViM2VwYmlrYzFkN20yIn0.pGuXQe015sVOtzEGrLsCGg";

export default {
  name: "MileStone",
  data() {
    return {
      version: null,
      center: [120.957283, 23.47],
      zoom: 14,
      map: null,
      gpxFile: null,
      gpxFileContent: null,
      progress: 0,
      previewed: false,
      layers: [],
      sources: [],
      markers: [],
      distance: "100",
      template: 'printf("%.1fK", dist/1000)',
      symbol: "Milestone",
      reverse: false,
      fits: false,
      terrainDistance: false,
    };
  },
  setup() {
    const { cookies } = useCookies();
    return { cookies };
  },
  mounted() {
    document.title = '里程產生器';
    this.loadCookies();
    this.getVersion();
    this.map = new mapboxgl.Map({
      container: "map", // container ID
      style: "mapbox://styles/mapbox/outdoors-v11", // style URL
      center: this.center, // starting position [lng, lat]
      zoom: this.zoom, // starting zoom
      projection: "globe", // display the map as a 3D globe
    });
    this.map.addControl(
      new MapboxLanguage({
        defaultLanguage: "zh-Hant",
      })
    );
  },
  methods: {
    onGpxChosen(event) {
      this.readGpxFile(event.target.files[0]);
    },
    onGpxDropped(event) {
      this.readGpxFile(event.dataTransfer.files[0]);
    },
    readGpxFile(gpxFile) {
      let reader = new FileReader();
      let self = this;
      reader.onload = function (event) {
        if (self.loadGpx(event.target.result)) {
          self.gpxFile = gpxFile;
          self.gpxFileContent = event.target.result;
        }
      };
      reader.readAsText(gpxFile);
    },
    loadGpx(gpx, fitBounds = true) {
      this.progress = 100;
      let parser = new DOMParser();
      let doc = parser.parseFromString(gpx, "application/xml");
      if (isDOMParseError(doc)) {
        alert("無法處理的GPX格式");
        this.progress = 0;
        return false;
      }
      this.clearMap();
      let geojson = toGeoJSON.gpx(doc);
      let colors = [
        "#ff0000",
        "#ff9900",
        "#ffff00",
        "#0cad00",
        "#00ffd5",
        "#00bbff",
        "#0040ff",
        "#d400ff",
        "#ff0077",
      ];
      let n = 0;
      for (let i = 0; i < geojson.features.length; i++) {
        let feature = geojson.features[i];
        switch (feature.geometry.type) {
          case "LineString":
          case "MultiLineString":
            if (!feature.properties) feature.properties = {};
            feature.properties.color = colors[n % colors.length];
            n++;
            break;
        }
      }
      let id = "gpx";
      this.map.addSource(id, {
        type: "geojson",
        data: geojson,
      });
      this.sources.push(id);
      id = "tracks";
      this.map.addLayer({
        id: id,
        type: "line",
        source: "gpx",
        layout: {
          "line-join": "round",
          "line-cap": "round",
        },
        paint: {
          "line-color": ["get", "color"],
          "line-width": 4,
        },
      });
      this.layers.push(id);
      let coordinates = [];
      for (let i = 0; i < geojson.features.length; i++) {
        let feature = geojson.features[i];
        switch (feature.geometry.type) {
          case "LineString": {
            coordinates = coordinates.concat(feature.geometry.coordinates);
            break;
          }
          case "MultiLineString": {
            for (let j = 0; j < feature.geometry.coordinates.length; j++) {
              coordinates = coordinates.concat(feature.geometry.coordinates[j]);
            }
            break;
          }
          case "Point": {
            let lngLat = [
              feature.geometry.coordinates[0],
              feature.geometry.coordinates[1],
            ];
            let options = null;
            let img = document.createElement("img");
            if (feature.properties.sym == this.symbol) {
              img.src = "/images/flag.svg";
              options = {
                element: img,
                anchor: "bottom-left",
              };
            } else {
              img.src = "/images/pin_drop.svg";
              options = {
                element: img,
                anchor: "bottom",
              };
            }
            let marker = new mapboxgl.Marker(options)
              .setLngLat(lngLat)
              .addTo(this.map);
            this.markers.push(marker);
            coordinates.push(lngLat);
            break;
          }
          default: {
            console.error("Unexpected feature type: " + feature.geometry.type);
            console.error(JSON.stringify(feature));
            break;
          }
        }
      }
      // for (let i = 0; i < coordinates.length; i++) {
      //   let c = coordinates[i];
      //   let lngLat = [c[0], c[1]];
      //   let options = null;
      //   let img = document.createElement("img");
      //   img.src = "/images/flag.svg";
      //   options = {
      //     element: img,
      //     anchor: 'bottom-left',
      //   };
      //   let marker = new mapboxgl.Marker(options)
      //     .setLngLat(lngLat)
      //     .addTo(this.map);
      // }
      id = "waypoints";
      this.map.addLayer({
        id: id,
        type: "symbol",
        source: "gpx",
        layout: {
          "text-field": ["get", "name"],
          "text-variable-anchor": ["left", "right", "bottom", "top"],
          "text-radial-offset": 0.5,
          "text-justify": "auto",
          "icon-image": ["get", "icon"],
        },
      });
      this.layers.push(id);
      if (fitBounds && coordinates.length > 0) {
        const bounds = new mapboxgl.LngLatBounds(
          coordinates[0],
          coordinates[0]
        );
        // Extend the 'LngLatBounds' to include every coordinate in the bounds result.
        for (const coord of coordinates) {
          bounds.extend(coord);
        }
        this.map.fitBounds(bounds, {
          padding: { top: 50, bottom: 150, left: 25, right: 25 },
        });
      }
      this.progress = 0;
      return true;
    },
    clear() {
      this.gpxFile = null;
      this.clearMap();
    },
    clearMap() {
      this.previewed = false;
      for (let i = 0; i < this.layers.length; i++) {
        this.map.removeLayer(this.layers[i]);
      }
      this.layers = [];
      for (let i = 0; i < this.sources.length; i++) {
        this.map.removeSource(this.sources[i]);
      }
      this.sources = [];
      for (let i = 0; i < this.markers.length; i++) {
        this.markers[i].remove();
      }
      this.markers = [];
    },
    getVersion() {
      let self = this;
      let xhr = new XMLHttpRequest();
      xhr.open("GET", "/cgi/version");
      xhr.onload = function () {
        if (xhr.status != 200) {
          alert(xhr.responseText);
        } else {
          let version = JSON.parse(xhr.response);
          self.version = version.tag ? version.tag : version.commit;
        }
      };
      xhr.send();
    },
    post(format, onresponse) {
      let self = this;
      let xhr = new XMLHttpRequest();
      let data = {
        format: format,
        distance: this.distance,
        template: this.template,
        symbol: this.symbol,
        reverse: this.reverse,
        fits: this.fits,
        terrainDistance: this.terrainDistance,
      };
      xhr.open("POST", "/cgi/milestones?" + encodeQueryData(data));
      xhr.setRequestHeader("Content-Type", "application/gpx+xml");
      xhr.onload = function () {
        self.progress = 0;
        if (xhr.status != 200) {
          alert(xhr.responseText);
        } else {
          onresponse(xhr.response, xhr.getResponseHeader("Content-Type"));
        }
      };
      xhr.send(this.gpxFile);
      this.progress = 100;
    },
    preview() {
      if (!this.distance) {
        alert("請輸入里程間距");
        return;
      }
      if (!this.template) {
        alert("請輸入里程航點名稱樣版");
        return;
      }
      if (!this.symbol) {
        alert("請輸入里程航點符號");
        return;
      }
      let self = this;
      this.post("gpx", function (response) {
        self.previewed = self.loadGpx(response, false);
        if (self.previewed) {
          self.saveCookies();
        }
      });
    },
    restore() {
      this.loadGpx(this.gpxFileContent, false);
    },
    download(format) {
      let filename =
        this.gpxFile.name.replace(/\.[^/.]+$/, "") + "(含里程)." + format;
      this.post(format, function (response, type) {
        let reader = new FileReader();
        reader.onload = function (event) {
          let a = document.createElement("a");
          a.href = event.target.result;
          a.download = filename;
          document.body.appendChild(a);
          a.click();
        };
        let blob = new Blob([response], { type: type });
        reader.readAsDataURL(blob);
      });
    },
    downloadGpx() {
      this.download("gpx");
    },
    downloadCsv() {
      this.download("csv");
    },
    saveCookies() {
      let parameters = {
        distance: this.distance,
        template: this.template,
        symbol: this.symbol,
        fits: this.fits,
        reverse: this.reverse,
        terrainDistance: this.terrainDistance,
      };
      this.cookies.set("milestone.parameters", JSON.stringify(parameters));
    },
    loadCookies() {
      let parameters = this.cookies.get("milestone.parameters");
      if (!parameters) return;
      this.distance = parameters.distance;
      this.template = parameters.template;
      this.symbol = parameters.symbol;
      this.fits = parameters.fits;
      this.reverse = parameters.reverse;
      this.terrainDistance = parameters.terrainDistance;
    },
  },
};
</script>
