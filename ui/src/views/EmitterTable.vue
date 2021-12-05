<template>
  <v-data-table
    :headers="headers"
    :items="items"
    class="elevation-1"
  >
    <template v-slot:top>
      <v-toolbar
        flat
      >
        <v-toolbar-title>Emitters</v-toolbar-title>
        <v-divider
          class="mx-4"
          inset
          vertical
        />
        <v-spacer />
        <v-dialog
          v-model="dialog"
          max-width="500px"
        >
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              color="primary"
              dark
              class="mb-2"
              v-bind="attrs"
              v-on="on"
            >
              New Emitter
            </v-btn>
          </template>
          <v-card>
            <v-card-title>
              <span class="text-h5">New Emitter</span>
            </v-card-title>

            <v-card-text>
              <v-container>
                <v-row>
                  <v-col cols="12">
                    <v-text-field
                      v-model="editedItem.name"
                      label="Name"
                      outlined
                      hide-details
                    />
                  </v-col>
                  <v-col cols="12">
                    <v-combobox
                      v-model="editedItem.type"
                      :items="availableTypeNames"
                      label="Emitter types"
                      outlined
                      hide-details
                    />
                  </v-col>
                </v-row>
                <rabbit-m-q-emitter-edit
                  v-if="editedItem.type == 'RabbitMQ'"
                  :item="editedItem"
                />
                <timescale-emitter-edit
                  v-if="editedItem.type == 'TimescaleDB'"
                  :item="editedItem"
                />
              </v-container>
            </v-card-text>

            <v-card-actions>
              <v-spacer />
              <v-btn
                color="blue darken-1"
                text
                @click="close"
              >
                Cancel
              </v-btn>
              <v-btn
                color="blue darken-1"
                text
                @click="save"
              >
                Save
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-toolbar>
    </template>
    <template v-slot:item.actions="{ item }">
      <v-icon
        class="mr-2"
        @click="editItem(item)"
      >
        mdi-pencil
      </v-icon>
      <v-icon
        @click="deleteItem(item)"
      >
        mdi-delete
      </v-icon>
    </template>
  </v-data-table>
</template>

<script>
  import ApiService from '@/services/api.service'

  export default {
    name: 'EmitterTable',

    data: () => ({
      dialog: false,
      dialogDelete: false,
      search: '',
      loading: false,
      headers: [
        {
          text: 'ID',
          align: 'start',
          filterable: false,
          value: 'ID',
          width: 75,
        },
        { text: 'Name', value: 'name', width: '20%' },
        { text: 'Type', value: 'type', width: '20%' },
        { text: 'Settings', value: 'settings', width: '60%' },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      items: [],
      availableTypeNames: [],
      editedIndex: -1,
      editedItem: {},
      defaultItem: {
        instance: {
          urls: [],
          channel: '',
        },
      },
      urls: [],
    }),

    created () {
      this.loading = true
      this.editedItem = Object.assign({}, this.defaultItem)
      this.editedIndex = -1
      ApiService.get('emitter')
        .then(response => {
          console.log('response.data: ' + JSON.stringify(response.data))
          this.items = response.data || []
          this.loading = false
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })

      ApiService.get('emitter/types')
        .then(response => {
          for (var i = 0; i < response.data.length; i++) {
            this.availableTypeNames.push(response.data[i])
          }
          console.log('available types: ' + this.availableTypeNames)
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
    },

    methods: {
      initialize () {},

      editItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item)
        this.dialog = true
      },

      deleteItem (item) {
        ApiService.delete('data/emitters/' + item.ID)
          .then(response => {
            for (var i = 0; i < this.items.length; i++) {
              if (this.items[i].ID === item.ID) this.items.splice(i, 1)
            }
            this.$notification.success('Emitter deleted')
          }).catch(response => {
            console.log('ERROR response: ' + response.message)
          })
      },

      close () {
        this.dialog = false
        this.$nextTick(() => {
          this.editedItem = Object.assign({}, this.defaultItem)
          this.editedIndex = -1
        })
      },

      save () {
        this.editedItem.settings = JSON.stringify(this.editedItem.instance)
        delete this.editedItem.instance
        var t = this
        if (this.editedIndex > -1) {
          Object.assign(this.items[this.editedIndex], this.editedItem)
          ApiService.put('emitter/' + this.editedItem.ID, this.editedItem)
            .then(response => {
              t.$notification.success('Emitter updated!')
            }).catch(function (response) {
              console.log('Failed to update emitter! ' + response)
              t.$notification.error('Failed to update emitter!' + response)
            })
        } else {
          ApiService.post('emitter/RabbitMQ', this.editedItem)
            .then(response => {
              t.$notification.success('Emitter created!')
              t.items.push(response.data)
            }).catch(function (response) {
              console.log('Failed to create emitter! ' + response.message)
              t.$notification.error('Failed to create emitter!' + response)
            })
        }
        this.editedItem = Object.assign({}, this.defaultItem)
        this.editedIndex = -1
        this.close()
      },
    },
  }
</script>
