<template>
  <div>
    <v-data-table
      :headers="headers"
      :items="items"
      :search="search"
      class="elevation-1"
    >
      <template v-slot:top>
        <v-toolbar
          flat
        >
          <v-toolbar-title>Tags</v-toolbar-title>
          <v-divider
            class="mx-4"
            inset
            vertical
          />
          <v-spacer />
          <v-text-field
            v-model="search"
            append-icon="mdi-magnify"
            label="Search"
            single-line
            hide-details
            clearable
          />
          <v-dialog
            v-model="dialog"
            max-width="500px"
          >
            <v-card>
              <v-card-title>
                <span class="text-h5">Tag</span>
              </v-card-title>

              <v-card-text>
                <v-container>
                  <v-row>
                    <v-col
                      cols="12"
                    >
                      <v-text-field
                        v-model="editedItem.name"
                        label="Name"
                        hide-details
                        readonly
                      />
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col
                      cols="12"
                    >
                      <v-combobox
                        v-model="editedItem.ut"
                        :items="availableUpdateTypes"
                        item-text="name"
                        item-value="value"
                        label="Update Type"
                        hide-details
                        outlined
                      />
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col
                      cols="12"
                    >
                      <v-text-field
                        v-if="editedItem.ut.value == 1"
                        v-model.number="editedItem.interval"
                        label="Interval"
                        hide-details
                        outlined
                      />
                      <v-text-field
                        v-if="editedItem.ut.value == 2"
                        v-model.number="editedItem.integratingdeadband"
                        label="Integrating Deadband"
                        hide-details
                        outlined
                      />
                    </v-col>
                  </v-row>
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
    <div>{{ message }}</div>
  </div>
</template>

<script>
  import ApiService from '@/services/api.service'
  import WebsocketService from '@/services/websocket.service'
  export default {
    name: 'TagTableView',

    data: () => ({
      dialog: false,
      dialogDelete: false,
      search: '',
      loading: false,
      headers: [
        // {
        //   text: 'ID',
        //   align: 'start',
        //   filterable: false,
        //   value: 'id',
        //   width: 75,
        // },
        { text: 'Name', value: 'name', width: '60%' },
        { text: 'Update Type', value: 'updatetype', width: '15%' },
        { text: 'Interval', value: 'interval', width: '10%' },
        { text: 'Integrating Deadband', value: 'integratingdeadband', width: '20%' },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      items: [],
      editedIndex: -1,
      editedItem: {
        fullname: '',
        email: '',
        ut: { value: 0, name: '' },
      },
      defaultItem: {
        fullname: '',
        email: '',
        ut: { value: 0, name: '' },
      },
      availableUpdateTypes: [{ value: 0, name: 'Pass thru' }, { value: 1, name: 'Interval' }, { value: 2, name: 'Deadband' }, { value: 3, name: 'Disabled' }],
      groups: [],
      groupsTable: {},
      message: 'kalle',
    }),

    created () {
      this.loading = true
      ApiService.get('proxy/point')
        .then(response => {
          this.items = response.data
          this.loading = false
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })

      var t = this
      WebsocketService.topic('data.point', function (topic, message) {
        t.message = JSON.stringify(message)
      })
      WebsocketService.topic('meta.message', function (topic, message) {
        // t.message = JSON.stringify(message)
      })
    },

    methods: {
      initialize () {},

      editItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item)
        this.editedItem.ut = this.availableUpdateTypes[this.editedItem.updatetype]
        console.log('editing item: ' + JSON.stringify(this.editedItem))
        this.dialog = true
      },

      deleteItem (item) {
        console.log('deleting item: ' + JSON.stringify(item))
        ApiService.delete('data/opc_tags/' + item.ID)
          .then(response => {
            for (var i = 0; i < this.items.length; i++) {
              if (this.items[i].ID === item.ID) this.items.splice(i, 1)
            }
            this.$notification.success('Tag deleted')
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
        if (this.editedIndex > -1) {
          this.editedItem.updatetype = this.editedItem.ut.value
          console.log('edited item: ' + JSON.stringify(this.editedItem))
          Object.assign(this.items[this.editedIndex], this.editedItem)
          ApiService.put('proxy/point', this.editedItem)
            .then(response => {
              console.log('updated item: ' + JSON.stringify(response.data.item))
              this.$notification.success('Point ' + response.data.item.name + ' successfully updated!')
            }).catch(response => {
              this.$notification.error('Failed to update point!' + response)
            })
        }
        this.close()
      },
    },
  }
</script>
