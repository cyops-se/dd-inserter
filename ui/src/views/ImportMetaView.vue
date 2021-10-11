<template>
  <div>
    <file-drop
      :dialog.sync="uploadDialog"
      :multiple="false"
      @filesUploaded="processUpload($event)"
    />
    <div v-if="showcontent">
      <v-data-table
        :headers="headers"
        :items="items"
        class="elevation-1"
      >
        <template v-slot:top>
          <v-toolbar
            flat
          >
            <v-toolbar-title>Tag meta - import changes</v-toolbar-title>
            <v-divider
              class="mx-4"
              inset
              vertical
            />
            <v-spacer />
            <v-dialog
              v-model="textDialog"
              max-width="80%"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-btn
                  color="primary"
                  dark
                  v-bind="attrs"
                  v-on="on"
                >
                  Import text
                </v-btn>
              </template>
              <v-card>
                <v-card-title>
                  <span class="text-h5">Paste content</span>
                </v-card-title>

                <v-card-text>
                  <v-container>
                    <v-row>
                      <v-col cols="12">
                        <v-textarea
                          v-model="content"
                          label="CSV meta content"
                          outlined
                          style="font-size: .75rem; line-height: initial"
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
                    @click="importText"
                  >
                    Import
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-dialog>
            <v-btn
              color="primary"
              dark
              class="ml-3"
              @click="uploadDialog = !uploadDialog"
            >
              Import file
            </v-btn>
            <v-btn
              color="success"
              dark
              class="ml-3"
              :disabled="saveDisabled"
              @click="saveChanges"
            >
              Save changes
            </v-btn>
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
    </div>
  </div>
</template>

<script>
  import FileDrop from '../components/app/FileDrop.vue'
  import ApiService from '@/services/api.service'
  export default {
    title: 'Import Meta View',
    components: { FileDrop },

    data: () => ({
      saveDisabled: true,
      textDialog: false,
      uploadDialog: false,
      showcontent: true,
      headers: [
        {
          text: 'ID',
          align: 'start',
          filterable: false,
          value: 'tag_id',
          width: 75,
        },
        { text: 'Name', value: 'name', width: '20%' },
        { text: 'Description', value: 'description', width: '60%' },
        { text: 'Unit', value: 'unit', width: '5%' },
        { text: 'Min', value: 'min', width: '5%' },
        { text: 'Max', value: 'max', width: '5%' },
        { text: 'Changed', value: 'changed', width: '5%' },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      items: [],
      loading: false,
      content: '',
    }),

    created () {
      ApiService.get('meta/all')
        .then(response => {
          this.items = response.data
          // for (var i = 0; i < this.items.length; i++) {
          //   this.items[i].changed = false
          // }
          this.loading = false
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
    },

    methods: {
      processUploadV1 (files) {
        var t = this
        ApiService.upload('meta/import', files)
          .then(response => {
            t.$notification.success('CSV uploaded')
            try {
              t.processResponse(response.data)
            } catch (error) {
              t.$notification.error('Processing response failed: ' + error.message)
            }
          }).catch(function (response) {
            t.$notification.error('Upload failed!' + response)
          })
      },

      processUpload (files) {
        var reader = new FileReader()
        var t = this
        reader.onload = function (event) {
          console.log('file content loaded: ' + event.target.result)
          var j = t.csvJSON(event.target.result)
          t.content = j
          t.processResponse(j)
        }
        console.log('loading file: ' + files[0].name)
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

        for (var mi = 0; mi < records.length; mi++) {
          var record = records[mi]
          var inuse = record[0]
          var tagname = record[1]
          var description = record[2]
          var min = parseFloat(record[3])
          var max = parseFloat(record[4])
          var unit = record[5]

          if (inuse !== 'x') continue
          // if (tagname !== 'A001092AQ900') continue

          for (var i = 0; i < this.items.length; i++) {
            var item = this.items[i]

            if (item.name.indexOf(tagname) === -1) continue
            var same = item.description === description
            if (same) same = item.unit === unit
            if (same) same |= item.min === min
            if (same) same |= item.max === max

            console.log('same: ' + same, item.description, description, item.description === description)
            if (!same) {
              item.description = description
              item.unit = unit
              item.min = min
              item.max = max
              item.changed = true
              console.log('item changed: ' + JSON.stringify(item))
            } else {
              item.changed = false
            }
            break
          }
        }

        // keep changed items in the table
        this.items = this.items.filter(item => item?.changed === true || false)

        this.saveDisabled = false
      },

      close () {
        this.textDialog = false
      },

      importText () {
        this.textDialog = false
        var j = this.csvJSON(this.content)
        this.processResponse(j)
      },

      saveChanges () {
        var t = this
        ApiService.post('meta/changes', this.items)
          .then(response => {
            t.$notification.success('Changes saved')
          }).catch(function (response) {
            t.$notification.error('Failed to save changes: ' + response)
          })
      },

      csvJSON (csv) {
        var lines = csv.split('\n')
        var result = []

        lines.map((line, indexLine) => {
          if (indexLine < 1) return // Jump header line
          var currentline = line.split(';')
          result.push(currentline)
        })

        // result.pop() // remove the last item because undefined values
        return result // JavaScript object
      },
    },
  }
</script>
