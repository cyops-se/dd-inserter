<template>
  <v-container>
    URLs
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
          v-model="url"
          label="URL"
          append-outer-icon="mdi-plus"
          @click:append-outer="addItem"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
  export default {
    name: 'ListEdit',

    props: {
      items: [],
    },

    data: () => ({
      url: '',
    }),

    methods: {
      addItem () {
        if (this.url === '') return
        this.items.push(this.url)
        this.url = ''
      },

      deleteItem (item, index) {
        console.log('deleting item: ' + JSON.stringify(item))
        this.url = item
        this.items.splice(index, 1)
      },
    },
  }
</script>
