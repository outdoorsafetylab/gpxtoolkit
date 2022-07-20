<template>
  <router-view />
</template>
<script>
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap";
import mapboxgl from "mapbox-gl";

mapboxgl.accessToken = "pk.eyJ1Ijoib3V0ZG9vcnNhZmV0eWxhYiIsImEiOiJjbDU1Y2N1eW0wbTViM2VwYmlrYzFkN20yIn0.pGuXQe015sVOtzEGrLsCGg";

export default {
  name: 'App',
  components: {}
}
</script>
<style>
</style>
