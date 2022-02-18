<template>
  <v-card
    v-bind="$attrs"
    class="v-card--material mt-4"
  >
    <v-card-title class="align-start">
      <v-sheet
        :color="color"
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
      color: 'success',
    }),

    created () {
      var t = this
      this.resetTimer()
      WebsocketService.topic('data.point', function (topic, message) {
        // t.emitter.count = JSON.stringify(message)
        if (message.n === 'Total Received') {
          t.emitter.count = message.v
          t.emitter.lastrun = message.t
          t.color = 'success'
          t.resetTimer()
        }
        if (message.n === 'Sequence Number') {
          t.emitter.seqno = message.v
          t.emitter.lastrun = message.t
          t.color = 'success'
          t.resetTimer()
        }
      })
    },

    methods: {
      resetTimer () {
        clearInterval(this.timer)
        this.timer = setInterval(this.invalidate, 10000)
      },

      invalidate () {
        this.color = 'error'
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
