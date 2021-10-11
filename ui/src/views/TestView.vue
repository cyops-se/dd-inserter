<template>
  <v-container
    id="logs-view"
    fluid
    tag="section"
  >
    <list-edit :items="items" />
  </v-container>
</template>

<script>
  export default {
    name: 'TestView',

    data: () => ({
      items: ['kalle anka', 'arne och hobbe'],
    }),

  }
</script>
