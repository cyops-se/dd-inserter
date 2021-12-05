<template>
  <v-container
    fluid
  >
    <v-form>
      <v-container>
        <v-row>
          <v-col
            cols="12"
          >
            <v-row>
              <v-col>
                <v-text-field
                  v-model="smtpserver.value"
                  label="SMTP Server"
                  placeholder="localhost"
                  filled
                />
              </v-col>
              <v-col>
                <list-edit
                  title="Recipients"
                  label="Email address"
                  :items="emails"
                />
              </v-col>
            </v-row>
          </v-col>
        </v-row>
        <hr class="my-2">
        <v-btn
          color="success"
          @click="save"
        >
          Save
        </v-btn>
        <v-btn
          color="warning"
          class="ml-4"
          @click="testMail"
        >
          Send test mail
        </v-btn>
      </v-container>
    </v-form>
  </v-container>
</template>

<script>
  // Utilities
  import ApiService from '@/services/api.service'

  export default {
    name: 'Monitoring',

    components: {
    },

    data: () => ({
      smtpserver: {},
      recipients: {},
      emails: [],
    }),

    computed: {
    },

    created () {
      ApiService.get('data/key_value_pairs/field/key/monitor.recipients')
        .then(response => {
          this.recipients = response.data[0]
          if (this.recipients.value !== '') {
            this.emails = this.recipients.value.split(',')
          }
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response.data))
        })
      ApiService.get('data/key_value_pairs/field/key/monitor.smtp')
        .then(response => {
          this.smtpserver = response.data[0]
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
    },

    methods: {
      save () {
        this.recipients.value = this.emails.join()
        ApiService.put('data/key_value_pairs', this.recipients)
          .then(response => {
            this.$notification.success('Recipients updated!')
          }).catch(response => {
            console.log('ERROR response: ' + JSON.stringify(response))
          })
        ApiService.put('data/key_value_pairs', this.smtpserver)
          .then(response => {
            this.$notification.success('SMTP server updated!')
          }).catch(response => {
            console.log('ERROR response: ' + JSON.stringify(response))
          })
      },

      testMail () {
        ApiService.post('system/testmail', this.emails.join())
          .then(response => {
            this.$notification.success('Test mail sent!')
          }).catch(response => {
            console.log('ERROR response: ' + JSON.stringify(response))
          })
      },
    },
  }
</script>
