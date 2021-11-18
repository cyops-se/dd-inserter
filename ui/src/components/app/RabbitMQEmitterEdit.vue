<template>
  <v-row>
    <v-col cols="12">
      <v-text-field
        v-model="item.instance.channel"
        label="Channel"
        outlined
        hide-details
      />
    </v-col>
    <v-col cols="12">
      <v-checkbox
        v-model="item.instance.durable"
        label="Durable"
        hide-details
        class="mt-n3"
        :value="item.instance.durable"
      />
    </v-col>
    <v-col cols="12">
      <list-edit :items="item.instance.urls" />
    </v-col>
  </v-row>
</template>

<script>
  export default {
    name: 'RabbitMQEmitterEdit',

    props: {
      item: Object,
    },
  }
</script>
