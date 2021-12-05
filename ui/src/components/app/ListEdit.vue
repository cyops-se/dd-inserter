<template>
  <v-card>
    <v-card-title>{{ title }}</v-card-title>
    <v-card-text>
      <v-list
        dense
      >
        <v-list-item
          v-for="(item, index) in items"
          :key="item.id"
          dense
        >
          <v-list-item-content>{{ item }}</v-list-item-content>
          <v-list-item-icon>
            <v-icon @click="deleteItem(item, index)">
              mdi-close
            </v-icon>
          </v-list-item-icon>
        </v-list-item>
      </v-list>
      <v-row>
        <v-col cols="12">
          <v-text-field
            v-model="item"
            :label="label"
            append-outer-icon="mdi-plus"
            @click:append-outer="addItem"
          />
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
</template>

<script>
  export default {
    name: 'ListEdit',

    props: {
      items: Array,
      title: { type: String, default: 'URLs' },
      label: { type: String, default: 'URL' },
    },

    data: () => ({
      item: '',
    }),

    methods: {
      addItem () {
        if (this.item === '') return
        this.items.push(this.item)
        this.item = ''
      },

      deleteItem (item, index) {
        this.item = item
        this.items.splice(index, 1)
      },
    },
  }
</script>
