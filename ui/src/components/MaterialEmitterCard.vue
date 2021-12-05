<template>
  <v-card
    v-bind="$attrs"
    class="v-card--material mt-4"
  >
    <v-card-title class="align-start">
      <v-sheet
        color="info"
        width="100%"
        class="overflow-hidden mt-n9 transition-swing v-card--material__sheet"
        elevation="6"
        max-width="100%"
        rounded
      >
        <v-theme-provider dark>
          <v-row align="center">
            <v-col cols="2">
              <div class="pa-5">
                <v-icon large>
                  mdi-tag-multiple
                </v-icon>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="pa-5 white--text">
                <span class="text-h3 text-no-wrap">
                  {{ emitter.name }}
                </span>
              </div>
            </v-col>
            <v-col cols="3">
              <div class="text-right nowrap">
                <div>Count: {{ emitter.count }}</div>
                <div>Sequence: {{ emitter.seqno }}</div>
              </div>
            </v-col>
          </v-row>
        </v-theme-provider>
      </v-sheet>

      <div class="pl-3 text-h4 v-card--material__title">
        <div class="text-subtitle-1 mb-n4 mt-4">
          <template>
            {{ emitter.description }}
          </template>
        </div>
      </div>
    </v-card-title>

    <slot />

    <template>
      <v-divider class="mt-2 mx-4" />

      <v-card-actions class="px-4 text-caption grey--text">
        <v-icon
          class="mr-1"
          small
        >
          mdi-clock-outline
        </v-icon>

        <span
          class="text-caption grey--text font-weight-light"
          v-text="'Last run: ' + emitter.lastrun.replace('T', ' ').substring(0, 19)"
        />
      </v-card-actions>
    </template>
  </v-card>
</template>

<script>
  import ApiService from '@/services/api.service'
  import WebsocketService from '@/services/websocket.service'
  export default {
    name: 'MaterialEmitterCard',

    inheritAttrs: false,

    props: {
      emitter: {
        type: Object,
        default: () => ({}),
      },
      eventHandlers: {
        type: Array,
        default: () => ([]),
      },
    },

    data: () => ({
      copy: {},
      timer: null,
    }),

    created () {
      var t = this
      WebsocketService.topic('data.point', function (topic, message) {
        // t.emitter.count = JSON.stringify(message)
        if (message.n === 'Total Received') {
          t.emitter.count = message.v
          t.emitter.lastrun = message.t
        }
        if (message.n === 'Sequence Number') {
          t.emitter.seqno = message.v
          t.emitter.lastrun = message.t
        }
      })
    },

    methods: {
      startStop () {
        var action = this.copy.status === 1 ? 'stop' : 'start'
        ApiService.get('opc/group/' + action + '/' + this.group.ID)
          .then(response => {
            this.$notification.success('Collection of group tags ' + (this.copy.status === 1 ? 'stopped' : 'started'))
            this.copy.status = this.copy.status === 1 ? 0 : 1
            if (this.copy.status === 1) {
              if (!this.timer) {
                clearInterval(this.timer)
                this.timer = setInterval(this.refresh, this.copy.interval * 1000)
              }
            } else {
              clearInterval(this.timer)
              this.timer = null
            }
          }).catch(response => {
            console.log('ERROR response: ' + response.message)
            this.$notification.error('Failed to start collection of group tags: ' + response.message)
          })
      },

      refresh () {
        ApiService.get('opc/group/' + this.group.ID)
          .then(response => {
            this.copy = response.data
          }).catch(response => {
            console.log('ERROR response (refresh): ' + response.message)
          })
      },
    },
  }
</script>

<style lang="sass">
.group-button
  font-size: .875rem !important
  margin-left: auto
  text-align: right

.nowrap
  white-space: nowrap
</style>
