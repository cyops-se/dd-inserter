<template>
  <v-container
    id="dashboard-view"
    fluid
    tag="section"
  >
    <v-row>
      <v-col cols="12">
        <v-row>
          <v-col
            v-for="(emitter, i) in emitters"
            :key="`emitter-${i}`"
            cols="12"
            md="6"
            lg="4"
          >
            <material-emitter-card :emitter="emitter" />
          </v-col>
        </v-row>
      </v-col>

      <v-col
        v-for="({ actionIcon, actionText, ...attrs }, i) in stats"
        :key="i"
        cols="12"
        md="6"
        lg="3"
      >
        <material-stat-card v-bind="attrs">
          <template #actions>
            <v-icon
              class="mr-2"
              small
              v-text="actionIcon"
            />
            <div class="text-truncate">
              {{ actionText }}
            </div>
          </template>
        </material-stat-card>
      </v-col>
      <error-logs-tables-view />
    </v-row>
  </v-container>
</template>

<script>
  // Utilities
  import ErrorLogsTablesView from './ErrorLogs'
  import ApiService from '@/services/api.service'

  export default {
    name: 'DashboardView',

    components: {
      ErrorLogsTablesView,
    },

    data: () => ({
      stats: [],
      // stats: [
      //   {
      //     actionIcon: 'mdi-calendar-range',
      //     actionText: 'Updated ',
      //     color: '#FD9A13',
      //     icon: 'mdi-tag',
      //     title: 'Tags',
      //     value: '',
      //   },
      //   {
      //     actionIcon: 'calendar-range',
      //     actionText: 'Updated',
      //     color: 'primary',
      //     icon: 'mdi-folder',
      //     title: 'emitters',
      //     value: '',
      //   },
      // ],
      tabs: 0,
      tags: [],
      emitters: [],
      servers: [],
    }),

    computed: {
    },

    created () {
      this.loading = true
      ApiService.get('data/emitters')
        .then(response => {
          this.emitters = response.data
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
    },
  }
</script>
