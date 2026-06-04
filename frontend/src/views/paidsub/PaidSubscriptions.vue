<template>
  <v-card flat>
    <v-card-title class="d-flex align-center">
      <v-icon class="mr-2">mdi-cash-multiple</v-icon>
      {{ $t('pages.paidSub') }}
      <v-chip class="ml-3" color="warning" size="small" variant="flat">experimental</v-chip>
      <v-spacer />
      <v-btn color="primary" :loading="loading" variant="tonal" @click="reloadAll">
        <v-icon start>mdi-refresh</v-icon>{{ $t('actions.refresh') ?? 'Refresh' }}
      </v-btn>
    </v-card-title>

    <v-alert
      v-if="!secretboxKeySet"
      type="warning"
      variant="tonal"
      class="mx-4 mb-2"
      density="comfortable"
    >
      For production, set the <strong>SUI_SECRETBOX_KEY</strong> environment variable so payment tokens
      are encrypted with a key kept outside the database.
    </v-alert>

    <v-tabs v-model="tab" color="primary" class="px-2">
      <v-tab value="bot">Bot</v-tab>
      <v-tab value="bindings">Bindings</v-tab>
      <v-tab value="autoreg">Auto-registration</v-tab>
      <v-tab value="tariffs">Tariffs</v-tab>
      <v-tab value="payments">Payments</v-tab>
      <v-tab value="orders">Orders</v-tab>
    </v-tabs>

    <v-window v-model="tab" class="pa-4">
      <!-- BOT -->
      <v-window-item value="bot">
        <v-row>
          <v-col cols="12" md="6">
            <v-switch v-model="enabled" color="primary" label="Enable client bot" hide-details />
          </v-col>
          <v-col cols="12" md="6">
            <v-text-field v-model="settings.paidSubBotPollSeconds" type="number" label="Long-poll timeout (s)" />
          </v-col>
          <v-col cols="12">
            <SettingsSecretField
              v-model="settings.paidSubBotToken"
              :has-secret="settings.paidSubBotTokenHasSecret"
              label="Bot token (separate from admin bot)"
            />
          </v-col>
        </v-row>
        <v-btn color="primary" :loading="loading" @click="saveSettings">{{ $t('actions.set') ?? 'Save' }}</v-btn>
      </v-window-item>

      <!-- BINDINGS -->
      <v-window-item value="bindings">
        <v-data-table :headers="bindingHeaders" :items="bindings" :loading="bindingsLoading" density="comfortable">
          <template #item.enable="{ item }">
            <v-chip :color="item.enable ? 'success' : 'error'" size="small" variant="flat">
              {{ item.enable ? 'active' : 'disabled' }}
            </v-chip>
          </template>
          <template #item.tgUserId="{ item }">
            <span v-if="item.tgUserId">{{ item.tgUserId }}</span>
            <span v-else class="text-disabled">—</span>
          </template>
          <template #item.actions="{ item }">
            <v-btn size="small" variant="text" icon="mdi-pencil" @click="openBinding(item)" />
            <v-btn v-if="item.tgUserId" size="small" variant="text" icon="mdi-link-off" color="error" @click="unbind(item)" />
          </template>
        </v-data-table>
      </v-window-item>

      <!-- AUTO-REGISTRATION -->
      <v-window-item value="autoreg">
        <v-row>
          <v-col cols="12" md="6">
            <v-switch v-model="autoRegister" color="primary" label="Auto-register unknown users" hide-details />
          </v-col>
          <v-col cols="12" md="6">
            <v-select
              v-model="autoInbounds"
              :items="inboundOptions"
              item-title="title"
              item-value="value"
              label="Inbounds for new clients"
              multiple
              chips
            />
          </v-col>
          <v-col cols="12" md="4">
            <v-text-field v-model="settings.paidSubTrialDays" type="number" label="Trial days" />
          </v-col>
          <v-col cols="12" md="4">
            <v-text-field v-model="settings.paidSubTrialVolumeGB" type="number" label="Trial traffic (GB, 0 = unlimited)" />
          </v-col>
          <v-col cols="12" md="4">
            <v-text-field v-model="settings.paidSubMaxClients" type="number" label="Max auto-registered clients" />
          </v-col>
          <v-col cols="12" md="4">
            <v-text-field v-model="settings.paidSubStartRateLimitPerMin" type="number" label="/start rate limit (per min)" />
          </v-col>
        </v-row>
        <v-btn color="primary" :loading="loading" @click="saveSettings">{{ $t('actions.set') ?? 'Save' }}</v-btn>
      </v-window-item>

      <!-- TARIFFS -->
      <v-window-item value="tariffs">
        <div class="d-flex mb-2">
          <v-spacer />
          <v-btn color="primary" @click="openTariff()"><v-icon start>mdi-plus</v-icon>Add tariff</v-btn>
        </div>
        <v-data-table :headers="tariffHeaders" :items="tariffs" :loading="tariffsLoading" density="comfortable">
          <template #item.price="{ item }">{{ (item.price / 100).toFixed(2) }} {{ item.currency }}</template>
          <template #item.starsAmount="{ item }">{{ item.starsAmount || '—' }}</template>
          <template #item.addTrafficBytes="{ item }">{{ item.addTrafficBytes ? (item.addTrafficBytes / (1024*1024*1024)).toFixed(2) + ' GB' : '∞' }}</template>
          <template #item.enabled="{ item }">
            <v-chip :color="item.enabled ? 'success' : 'grey'" size="small" variant="flat">{{ item.enabled ? 'on' : 'off' }}</v-chip>
          </template>
          <template #item.actions="{ item }">
            <v-btn size="small" variant="text" icon="mdi-pencil" @click="openTariff(item)" />
            <v-btn size="small" variant="text" icon="mdi-delete" color="error" @click="deleteTariff(item)" />
          </template>
        </v-data-table>
      </v-window-item>

      <!-- PAYMENTS -->
      <v-window-item value="payments">
        <v-row>
          <v-col cols="12" md="4">
            <v-text-field v-model="settings.paidSubCurrency" label="Default currency (e.g. RUB, USD)" maxlength="3" />
          </v-col>
          <v-col cols="12" md="4">
            <v-text-field v-model="settings.paidSubOrderTTLMinutes" type="number" label="Pending order TTL (min)" />
          </v-col>
        </v-row>
        <v-divider class="my-2" />
        <v-switch v-model="starsEnabled" color="primary" label="Telegram Stars (XTR)" hide-details />
        <v-divider class="my-2" />
        <v-switch v-model="yooEnabled" color="primary" label="YooKassa" hide-details />
        <SettingsSecretField
          v-model="settings.paidSubYooKassaToken"
          :has-secret="settings.paidSubYooKassaTokenHasSecret"
          label="YooKassa provider_token (BotFather)"
        />
        <v-divider class="my-2" />
        <v-switch v-model="stripeEnabled" color="primary" label="Stripe" hide-details />
        <SettingsSecretField
          v-model="settings.paidSubStripeToken"
          :has-secret="settings.paidSubStripeTokenHasSecret"
          label="Stripe provider_token (BotFather)"
        />
        <v-divider class="my-2" />
        <v-switch v-model="cryptoEnabled" color="primary" label="CryptoBot" hide-details />
        <SettingsSecretField
          v-model="settings.paidSubCryptoBotToken"
          :has-secret="settings.paidSubCryptoBotTokenHasSecret"
          label="CryptoBot API token"
        />
        <v-divider class="my-2" />
        <v-switch v-model="externalEnabled" color="primary" label="External payment link" hide-details />
        <v-text-field
          v-model="settings.paidSubExternalUrlTemplate"
          label="External URL template (https://… with {orderId} {amount} {currency} {clientId})"
        />
        <v-btn class="mt-2" color="primary" :loading="loading" @click="saveSettings">{{ $t('actions.set') ?? 'Save' }}</v-btn>
      </v-window-item>

      <!-- ORDERS -->
      <v-window-item value="orders">
        <v-data-table :headers="orderHeaders" :items="orders" :loading="ordersLoading" density="comfortable">
          <template #item.amount="{ item }">{{ (item.amount / 100).toFixed(2) }} {{ item.currency }}</template>
          <template #item.status="{ item }">
            <v-chip :color="orderStatusColor(item.status)" size="small" variant="flat">{{ item.status }}</v-chip>
          </template>
          <template #item.createdAt="{ item }">{{ item.createdAt ? new Date(item.createdAt * 1000).toLocaleString() : '' }}</template>
        </v-data-table>
      </v-window-item>
    </v-window>
  </v-card>

  <!-- Binding dialog -->
  <v-dialog v-model="bindingDialog" max-width="420">
    <v-card>
      <v-card-title>Bind Telegram to {{ bindingEdit.name }}</v-card-title>
      <v-card-text>
        <v-text-field v-model="bindingEdit.tgUserId" type="number" label="Telegram user ID (0 = unbind)" autofocus />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="bindingDialog = false">{{ $t('actions.cancel') ?? 'Cancel' }}</v-btn>
        <v-btn color="primary" @click="saveBinding">{{ $t('actions.set') ?? 'Save' }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Tariff dialog -->
  <v-dialog v-model="tariffDialog" max-width="560">
    <v-card>
      <v-card-title>{{ tariffEdit.id ? 'Edit tariff' : 'New tariff' }}</v-card-title>
      <v-card-text>
        <v-text-field v-model="tariffEdit.name" label="Name" />
        <v-text-field v-model="tariffEdit.description" label="Description" />
        <v-row>
          <v-col cols="6"><v-text-field v-model.number="tariffEdit.priceMajor" type="number" label="Price (major units)" /></v-col>
          <v-col cols="6"><v-text-field v-model="tariffEdit.currency" label="Currency" maxlength="3" /></v-col>
          <v-col cols="6"><v-text-field v-model.number="tariffEdit.starsAmount" type="number" label="Stars amount (XTR)" /></v-col>
          <v-col cols="6"><v-text-field v-model.number="tariffEdit.addDays" type="number" label="+Days" /></v-col>
          <v-col cols="6"><v-text-field v-model.number="tariffEdit.addTrafficGB" type="number" label="+Traffic (GB, 0 = unlimited)" /></v-col>
          <v-col cols="6"><v-text-field v-model.number="tariffEdit.sort" type="number" label="Sort" /></v-col>
        </v-row>
        <v-switch v-model="tariffEdit.enabled" color="primary" label="Enabled" hide-details />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="tariffDialog = false">{{ $t('actions.cancel') ?? 'Cancel' }}</v-btn>
        <v-btn color="primary" @click="saveTariff">{{ $t('actions.set') ?? 'Save' }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import HttpUtils from '@/plugins/httputil'
import SettingsSecretField from '@/components/SettingsSecretField.vue'
import { normalizeSecretFields, stripSecretPlaceholders } from '@/components/settingsSecretField'
import { push } from 'notivue'
import { i18n } from '@/locales'

type SMap = Record<string, string>

const SECRET_KEYS = ['paidSubBotToken', 'paidSubYooKassaToken', 'paidSubStripeToken', 'paidSubCryptoBotToken']

const defaults: SMap = {
  paidSubEnabled: 'false',
  paidSubBotToken: '',
  paidSubBotTokenHasSecret: 'false',
  paidSubBotPollSeconds: '25',
  paidSubAutoRegister: 'false',
  paidSubAutoInbounds: '[]',
  paidSubTrialDays: '3',
  paidSubTrialVolumeGB: '0',
  paidSubMaxClients: '5000',
  paidSubStartRateLimitPerMin: '3',
  paidSubCurrency: 'RUB',
  paidSubStarsEnabled: 'false',
  paidSubYooKassaEnabled: 'false',
  paidSubYooKassaToken: '',
  paidSubYooKassaTokenHasSecret: 'false',
  paidSubStripeEnabled: 'false',
  paidSubStripeToken: '',
  paidSubStripeTokenHasSecret: 'false',
  paidSubCryptoBotEnabled: 'false',
  paidSubCryptoBotToken: '',
  paidSubCryptoBotTokenHasSecret: 'false',
  paidSubExternalEnabled: 'false',
  paidSubExternalUrlTemplate: '',
  paidSubOrderTTLMinutes: '30',
}

const tab = ref('bot')
const loading = ref(false)
const settings = ref<SMap>({ ...defaults })
const secretboxKeySet = ref(true)

const pickSettings = (all: SMap): SMap => {
  const out: SMap = {}
  for (const k of Object.keys(defaults)) {
    if (all[k] !== undefined) out[k] = String(all[k])
    else out[k] = defaults[k]
  }
  return out
}

const boolSetting = (key: string) => computed({
  get: () => settings.value[key] === 'true',
  set: (v: boolean) => { settings.value[key] = v ? 'true' : 'false' },
})
const enabled = boolSetting('paidSubEnabled')
const autoRegister = boolSetting('paidSubAutoRegister')
const starsEnabled = boolSetting('paidSubStarsEnabled')
const yooEnabled = boolSetting('paidSubYooKassaEnabled')
const stripeEnabled = boolSetting('paidSubStripeEnabled')
const cryptoEnabled = boolSetting('paidSubCryptoBotEnabled')
const externalEnabled = boolSetting('paidSubExternalEnabled')

const autoInbounds = computed<number[]>({
  get: () => {
    try { return JSON.parse(settings.value.paidSubAutoInbounds || '[]') } catch { return [] }
  },
  set: (v: number[]) => { settings.value.paidSubAutoInbounds = JSON.stringify(v) },
})

const loadSettings = async () => {
  const msg = await HttpUtils.get('api/settings')
  if (msg.success) {
    const normalized = normalizeSecretFields({ ...defaults, ...(msg.obj ?? {}) }) as SMap
    settings.value = pickSettings(normalized)
  }
}

const loadStatus = async () => {
  const msg = await HttpUtils.get('api/paidsub/status')
  if (msg.success) secretboxKeySet.value = !!msg.obj?.secretboxKeySet
}

const saveSettings = async () => {
  loading.value = true
  const payload = stripSecretPlaceholders(pickSettings(settings.value)) as SMap
  const msg = await HttpUtils.post('api/save', { object: 'settings', action: 'set', data: JSON.stringify(payload) })
  if (msg.success) {
    push.success({ title: i18n.global.t('success'), message: i18n.global.t('pages.paidSub'), duration: 4000 })
    if (msg.obj?.settings) {
      const normalized = normalizeSecretFields({ ...defaults, ...msg.obj.settings }) as SMap
      settings.value = pickSettings(normalized)
    }
  }
  loading.value = false
}

// ---- inbounds for the auto-reg selector ----
const inboundOptions = ref<{ title: string; value: number }[]>([])
const loadInbounds = async () => {
  const msg = await HttpUtils.get('api/inbounds')
  if (msg.success && Array.isArray(msg.obj)) {
    inboundOptions.value = msg.obj.map((i: any) => ({ title: `${i.tag} (${i.type})`, value: i.id }))
  }
}

// ---- bindings ----
const bindings = ref<any[]>([])
const bindingsLoading = ref(false)
const bindingHeaders = [
  { title: 'Client', key: 'name' },
  { title: 'Telegram ID', key: 'tgUserId' },
  { title: 'Status', key: 'enable' },
  { title: '', key: 'actions', sortable: false, align: 'end' as const },
]
const bindingDialog = ref(false)
const bindingEdit = ref<{ clientId: number; name: string; tgUserId: number | string }>({ clientId: 0, name: '', tgUserId: 0 })

const loadBindings = async () => {
  bindingsLoading.value = true
  const msg = await HttpUtils.get('api/paidsub/bindings')
  if (msg.success) bindings.value = msg.obj ?? []
  bindingsLoading.value = false
}
const openBinding = (item: any) => {
  bindingEdit.value = { clientId: item.clientId, name: item.name, tgUserId: item.tgUserId || '' }
  bindingDialog.value = true
}
const saveBinding = async () => {
  const tgUserId = Number(bindingEdit.value.tgUserId) || 0
  const msg = await HttpUtils.post('api/paidsub/bindings', { clientId: bindingEdit.value.clientId, tgUserId })
  if (msg.success) { bindingDialog.value = false; await loadBindings() }
}
const unbind = async (item: any) => {
  const msg = await HttpUtils.post('api/paidsub/bindings', { clientId: item.clientId, tgUserId: 0 })
  if (msg.success) await loadBindings()
}

// ---- tariffs ----
const tariffs = ref<any[]>([])
const tariffsLoading = ref(false)
const tariffHeaders = [
  { title: 'Name', key: 'name' },
  { title: 'Price', key: 'price' },
  { title: 'Stars', key: 'starsAmount' },
  { title: '+Days', key: 'addDays' },
  { title: '+Traffic', key: 'addTrafficBytes' },
  { title: 'Enabled', key: 'enabled' },
  { title: '', key: 'actions', sortable: false, align: 'end' as const },
]
const tariffDialog = ref(false)
const blankTariff = () => ({ id: 0, name: '', description: '', priceMajor: 0, currency: settings.value.paidSubCurrency || 'RUB', starsAmount: 0, addDays: 30, addTrafficGB: 0, sort: 0, enabled: true })
const tariffEdit = ref<any>(blankTariff())

const loadTariffs = async () => {
  tariffsLoading.value = true
  const msg = await HttpUtils.get('api/paidsub/tariffs')
  if (msg.success) tariffs.value = msg.obj ?? []
  tariffsLoading.value = false
}
const openTariff = (item?: any) => {
  if (item) {
    tariffEdit.value = {
      id: item.id, name: item.name, description: item.description,
      priceMajor: (item.price || 0) / 100, currency: item.currency,
      starsAmount: item.starsAmount || 0, addDays: item.addDays || 0,
      addTrafficGB: (item.addTrafficBytes || 0) / (1024 * 1024 * 1024),
      sort: item.sort || 0, enabled: !!item.enabled,
    }
  } else {
    tariffEdit.value = blankTariff()
  }
  tariffDialog.value = true
}
const saveTariff = async () => {
  const e = tariffEdit.value
  const data: any = {
    name: e.name, description: e.description,
    price: Math.round(Number(e.priceMajor) * 100),
    currency: (e.currency || 'RUB').toUpperCase(),
    starsAmount: Math.round(Number(e.starsAmount) || 0),
    addDays: Math.round(Number(e.addDays) || 0),
    addTrafficBytes: Math.round((Number(e.addTrafficGB) || 0) * 1024 * 1024 * 1024),
    sort: Math.round(Number(e.sort) || 0),
    enabled: !!e.enabled,
  }
  const action = e.id ? 'edit' : 'new'
  if (e.id) data.id = e.id
  const msg = await HttpUtils.post('api/paidsub/tariffs', { action, data })
  if (msg.success) { tariffDialog.value = false; await loadTariffs() }
}
const deleteTariff = async (item: any) => {
  const msg = await HttpUtils.post('api/paidsub/tariffs', { action: 'del', data: item.id })
  if (msg.success) await loadTariffs()
}

// ---- orders ----
const orders = ref<any[]>([])
const ordersLoading = ref(false)
const orderHeaders = [
  { title: 'ID', key: 'id' },
  { title: 'Client', key: 'clientId' },
  { title: 'Provider', key: 'provider' },
  { title: 'Amount', key: 'amount' },
  { title: 'Status', key: 'status' },
  { title: 'Created', key: 'createdAt' },
]
const loadOrders = async () => {
  ordersLoading.value = true
  const msg = await HttpUtils.get('api/paidsub/orders')
  if (msg.success) orders.value = msg.obj ?? []
  ordersLoading.value = false
}
const orderStatusColor = (s: string) => ({ paid: 'success', pending: 'warning', failed: 'error', expired: 'grey', canceled: 'grey' } as any)[s] || 'grey'

const reloadAll = async () => {
  loading.value = true
  await Promise.all([loadSettings(), loadStatus(), loadInbounds(), loadBindings(), loadTariffs(), loadOrders()])
  loading.value = false
}

onMounted(reloadAll)
</script>
