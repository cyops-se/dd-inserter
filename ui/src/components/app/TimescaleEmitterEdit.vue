<template>
  <!-- <v-dialog
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
        New Group
      </v-btn>
    </template>
    <v-card>
      <v-card-title>
        <span class="text-h5">Group</span>
      </v-card-title>

      <v-card-text>
        <v-container>
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model="editedItem.name"
                label="Name"
                outlined
                hide-details
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model.number="editedItem.interval"
                label="Sampling Interval"
                type="number"
                outlined
                hide-details
              />
            </v-col>
            <v-col cols="12">
              <v-combobox
                v-model="editedItem.progid"
                :items="availableProgids"
                label="Server ProgID"
                outlined
                hide-details
              />
            </v-col>
            <v-col cols="12">
              <v-checkbox
                v-model="editedItem.runatstart"
                label="Start automatically"
                hide-details
                class="mt-n3"
                :value="editedItem ? editedItem.runatstart : true"
              />
            </v-col>
            <v-col cols="8">
              <v-text-field
                v-model.number="editedItem.targetip"
                label="Target IP (data diode)"
                outlined
                hide-details
              />
            </v-col>
            <v-col cols="4">
              <v-text-field
                v-model.number="editedItem.targetport"
                label="Target Port"
                outlined
                hide-details
              />
            </v-col>
            <v-col cols="12">
              <v-textarea
                v-model="editedItem.description"
                label="Description"
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
  </v-dialog> -->
  <h1>KALLE</h1>
</template>

<script>
  export default {
    name: 'MaterialCard',
  }
</script>
