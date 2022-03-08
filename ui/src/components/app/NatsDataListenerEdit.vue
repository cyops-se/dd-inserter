<template>
  <v-row>
    <v-col cols="12">
      <v-text-field
        v-model="item.instance.url"
        label="URL"
        outlined
        hide-details
      />
    </v-col>
    <v-col cols="12">
      <v-text-field
        v-model="item.instance.subject"
        label="Subject"
        outlined
        hide-details
      />
    </v-col>
  </v-row>
</template>

<script>
  export default {
    name: 'NatsDataListenerEdit',

    props: {
      item: Object,
    },
  }
</script>
