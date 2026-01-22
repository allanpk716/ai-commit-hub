<script setup lang="ts">
import { ref } from 'vue'
import { Greet } from '../wailsjs/go/main/App'

const message = ref('Click the button to greet!')

const greet = async () => {
  const result = await Greet('World')
  message.value = result
}
</script>

<template>
  <div class="app">
    <h1>AI Commit Hub</h1>
    <p>{{ message }}</p>
    <button @click="greet">Greet</button>
  </div>
</template>

<style scoped>
.app {
  padding: 20px;
}
</style>
