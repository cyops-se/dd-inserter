<template>
  <div>
    <file-drop
      :dialog.sync="uploadDialog"
      :multiple="false"
      @filesUploaded="processUpload($event)"
    />
    <v-data-table
      v-model="selected"
      :headers="headers"
      :items="items"
      :search="search"
      show-select
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
          <v-btn
            color="secondary"
            class="ml-2"
            :disabled="selected.length === 0"
            @click="editSelectedItems"
          >
            Edit selected
          </v-btn>
          <v-btn
            color="primary"
            class="ml-2"
            @click="exportCSV"
          >
            Export
          </v-btn>
          <v-btn
            color="primary"
            class="ml-2"
            @click="uploadDialog = !uploadDialog"
          >
            Import
          </v-btn>
          <v-btn
            color="success"
            class="ml-2"
            :disabled="saveChangedDisabled"
            @click="saveChanged"
          >
            Save changes
          </v-btn>
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
                        v-if="editedIndex != -2"
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
                      <v-text-field
                        v-if="editedIndex != -2"
                        v-model="editedItem.alias"
                        label="Alias"
                        hide-details
                      />
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col
                      cols="4"
                    >
                      <v-text-field
                        v-model.number="editedItem.min"
                        label="Min"
                        hide-details
                        outlined
                      />
                    </v-col>
                    <v-col
                      cols="4"
                    >
                      <v-text-field
                        v-model.number="editedItem.max"
                        label="Max"
                        hide-details
                        outlined
                      />
                    </v-col>
                    <v-col
                      cols="4"
                    >
                      <v-text-field
                        v-model.number="editedItem.engunit"
                        label="Engineering Unit"
                        hide-details
                        outlined
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
                        v-if="editedItem.ut === 'Interval'"
                        v-model.number="editedItem.interval"
                        label="Interval"
                        hide-details
                        outlined
                      />
                      <v-text-field
                        v-if="editedItem.ut === 'Deadband'"
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
  import Vue from 'vue'
  import ApiService from '@/services/api.service'
  import WebsocketService from '@/services/websocket.service'
  export default {
    name: 'TagTableView',

    data: () => ({
      dialog: false,
      dialogDelete: false,
      uploadDialog: false,
      search: '',
      loading: false,
      saveSelectedDisabled: true,
      saveChangedDisabled: true,
      content: '',
      headers: [
        {
          text: 'ID',
          align: 'start',
          filterable: false,
          value: 'id',
          width: 75,
        },
        { text: 'Name/Alias', value: 'dname', width: '20%' },
        { text: 'Description', value: 'description', width: '40%' },
        { text: 'Value', value: 'value', width: '10%' },
        { text: 'Update Type', value: 'updatetype', width: '5%' },
        { text: 'Interval', value: 'interval', width: '5%' },
        { text: 'Integrating Deadband', value: 'integratingdeadband', width: '5%' },
        { text: 'T', value: 'threshold', width: '5%' },
        { text: 'I', value: 'integrator', width: '5%' },
        { text: 'R', value: 'raw', width: '5%' },
        { text: 'Min', value: 'min', width: '5%' },
        { text: 'Max', value: 'max', width: '5%' },
        { text: 'Unit', value: 'engunit', width: '5%' },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      items: [],
      editedIndex: -1,
      editedItem: {
        ut: '',
        interval: 0,
        integratingdeadband: 0.3,
        min: 0,
        max: 100,
        engunit: '',
      },
      defaultItem: {
        ut: '',
        interval: 0,
        integratingdeadband: 0.3,
        min: 0,
        max: 100,
        engunit: '',
      },
      availableUpdateTypes: ['Pass thru', 'Interval', 'Deadband', 'Disabled'],
      groups: [],
      groupsTable: {},
      message: '',
      selected: [],
    }),

    created () {
      var t = this
      WebsocketService.topic('data.point', function (topic, message) {
        // t.message = JSON.stringify(message)
        var item = t.items.find(i => i.dname === message.n)
        if (item) Vue.set(item, 'value', parseFloat(message.v).toFixed(2))
      })
      WebsocketService.topic('meta.entry', function (topic, message) {
        // t.message = JSON.stringify(message)
        var item = t.items.find(i => i.dname === message.datapoint.n)
        if (item) {
          Vue.set(item, 'integrator', parseFloat(message.integrator).toFixed(2))
          Vue.set(item, 'threshold', parseFloat(message.threshold).toFixed(2))
          Vue.set(item, 'raw', parseFloat(message.datapoint.v).toFixed(2))
        }
      })

      this.refresh()
    },

    methods: {
      refresh () {
        ApiService.get('proxy/point')
          .then(response => {
            this.items = response.data
            this.items.forEach(i => i.dname = i.alias || i.name)
          }).catch(response => {
            console.log('ERROR response: ' + JSON.stringify(response))
          })
      },

      editItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item)
        this.editedItem.ut = this.availableUpdateTypes[this.editedItem.updatetype]
        console.log('editing item: ' + JSON.stringify(this.editedItem))
        this.dialog = true
      },

      editSelectedItems () {
        this.editedIndex = -2
        this.editedItem = Object.assign({}, this.defaultItem)
        this.dialog = true
      },

      deleteItem (item) {
        ApiService.delete('proxy/point/' + item.id)
          .then(response => {
            for (var i = 0; i < this.items.length; i++) {
              if (this.items[i].id === item.id) this.items.splice(i, 1)
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

      updateType (typename) {
        for (var i = 0; i < this.availableUpdateTypes.length; i++) {
          if (typename === this.availableUpdateTypes[i]) return i
        }
        return 0
      },

      save () {
        if (this.editedIndex > -1) {
          this.editedItem.updatetype = this.updateType(this.editedItem.ut)
          Object.assign(this.items[this.editedIndex], this.editedItem)
          ApiService.put('proxy/point', this.editedItem)
            .then(response => {
              this.$notification.success(response.data.item.name + ' successfully updated!')
              this.refresh()
            }).catch(response => {
              this.$notification.error('Failed to update point!' + response)
            })
        } else if (this.selected.length > 0) {
          for (var i = 0; i < this.selected.length; i++) {
            this.selected[i].updatetype = this.updateType(this.editedItem.ut)
            this.selected[i].interval = this.editedItem.interval
            this.selected[i].integratingdeadband = this.editedItem.integratingdeadband
            this.selected[i].min = this.editedItem.min
            this.selected[i].max = this.editedItem.max
            this.selected[i].engunit = this.editedItem.engunit
            ApiService.put('proxy/point', this.selected[i])
              .then(response => {
                // this.$notification.success(response.data.item.name + ' successfully updated!')
                this.refresh()
              }).catch(response => {
                this.$notification.error('Failed to update point!' + response)
              })
          }
        }
        this.close()
      },

      exportCSV () {
        let csvContent = 'data:text/csv;charset=utf-8,'

        csvContent += [
          'inuse;name;description;min;max;unit;type;interval;integratingdeadband;',
          ...this.items.map(item => 'x;' + item.name + ';' + item.description + ';' + parseFloat(item.min) + ';' + parseFloat(item.max) + ';' + item.engunit + ';' + parseInt(item.updatetype) + ';' + parseInt(item.interval) + ';' + parseFloat(item.integratingdeadband) + ';'),
        ]
          .join('\n')
          .replace(/(^\[)|(\]$)/gm, '')

        const data = encodeURI(csvContent)
        const link = document.createElement('a')
        link.setAttribute('href', data)
        link.setAttribute('download', 'export.csv')
        link.click()
      },

      processUpload (files) {
        var reader = new FileReader()
        var t = this
        reader.onload = function (event) {
          var j = t.csvJSON(event.target.result)
          t.content = j
          t.processResponse(j)
        }
        reader.readAsText(files[0])
      },

      processResponse (records) {
        // iterate through all existing items and compare content
        // assume the following column format:
        // col 0: x indicates in use
        // col 1: tag name
        // col 2: tag description
        // col 3: tag min value
        // col 4: tag max value
        // col 5: tag unit
        // col 6: update type (0 = pass thru, 1 = interval, 2 = integrating deadband, 3 = disabled)
        // col 7: interval value
        // col 8: integrating deadband value
        // col 9: alias to be used instead of name if specified

        for (var mi = 0; mi < records.length; mi++) {
          var record = records[mi]
          var inuse = record[0]
          var tagname = record[1]
          var description = record[2]
          var min = parseFloat(record[3])
          var max = parseFloat(record[4])
          var engunit = record[5]
          var updatetype = parseInt(record[6])
          var interval = parseInt(record[7])
          var integratingdeadband = parseFloat(record[8])
          var alias = record[9]

          if (inuse !== 'x') continue

          var found = false
          for (var i = 0; i < this.items.length; i++) {
            var item = this.items[i]

            if (item.name !== tagname) continue
            found = true
            var same = item.description === description
            if (same) same = item.engunit === engunit
            if (same) same = item.min === min
            if (same) same = item.max === max
            if (same) same = item.updatetype === updatetype
            if (same && updatetype === 1) same = item.interval === interval
            if (same && updatetype === 2) same = item.integratingdeadband === integratingdeadband
            if (same) same = item.alias === alias

            if (!same) {
              item.description = description
              item.engunit = engunit
              item.min = min
              item.max = max
              item.updatetype = updatetype
              if (updatetype === 1) item.interval = interval
              if (updatetype === 2) item.integratingdeadband = integratingdeadband
              item.alias = alias
              item.changed = true
            } else {
              item.changed = false
            }
            break
          }

          if (!found) {
            var newitem = { name: tagname, description: description, engunit: engunit, min: min, max: max, updatetype: updatetype, interval: interval, integratingdeadband: integratingdeadband, alias: alias, new: true }
            this.items.push(newitem)
          }
        }

        // console.log('all items: ' + JSON.stringify(this.items))
        // keep changed items in the table
        this.items = this.items.filter(item => (item?.changed === true || item?.new === true) || false)
        this.items.forEach(i => i.dname = i.alias || i.name)

        // console.log('changed items: ' + JSON.stringify(this.items))

        if (this.items.length > 0) this.saveChangedDisabled = false
      },

      csvJSON (csv) {
        var lines = csv.split('\n')
        var result = []

        lines.map((line, indexLine) => {
          if (indexLine < 1) return // Skip header line
          var currentline = line.split(';')
          result.push(currentline)
        })

        // result.pop() // remove the last item because undefined values
        return result // JavaScript object
      },

      saveChanged () {
        this.saveChangedDisabled = true
        var t = this
        console.log('items: ' + JSON.stringify(this.items))
        ApiService.post('proxy/points', this.items)
          .then(response => {
            t.$notification.success('Changes saved')
            t.refresh()
          }).catch(function (response) {
            t.$notification.error('Failed to save changes: ' + response)
          })
        console.log('save changes from import ...')
      },
    },
  }
</script>
