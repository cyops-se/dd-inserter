<template>
  <v-row>
    <v-col cols="6">
      <v-text-field
        v-model="item.instance.host"
        label="Host"
        outlined
        hide-details
      />
    </v-col>
    <v-col cols="6">
      <v-text-field
        v-model.number="item.instance.port"
        label="Port"
        outlined
        hide-details
      />
    </v-col>
    <v-col cols="6">
      <v-text-field
        v-model="item.instance.user"
        label="User"
        outlined
        hide-details
      />
    </v-col>
    <v-col cols="6">
      <v-text-field
        v-model="item.instance.password"
        label="Password"
        outlined
        hide-details
      />
    </v-col>
    <v-col cols="6">
      <v-text-field
        v-model="item.instance.database"
        label="Database"
        outlined
        hide-details
      />
    </v-col>
    <v-col cols="6">
      <v-text-field
        v-model.number="item.instance.batchsize"
        label="Batch size"
        outlined
        hide-details
      />
    </v-col>
  </v-row>
</template>

<script>
  export default {
    name: 'TimescaleEmitterEdit',

    props: {
      item: {},
    },
  }
</script>
