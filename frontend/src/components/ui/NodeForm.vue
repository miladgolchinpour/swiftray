<template>
  <div class="space-y-3.5">
    <!-- Name -->
    <FormField label="Node Name">
      <input
        :value="model.name"
        type="text"
        placeholder="My Node"
        class="field-input"
        @input="update('name', ($event.target as HTMLInputElement).value)"
      />
    </FormField>

    <!-- Protocol -->
    <div class="space-y-1.5">
      <label class="text-xs font-medium text-muted-foreground">Protocol</label>
      <div class="grid grid-cols-4 gap-1 p-0.5 rounded-lg bg-muted">
        <button
          v-for="p in protocols"
          :key="p.value"
          class="h-8 rounded-md text-xs font-medium transition-all duration-150"
          :class="model.protocolType === p.value
            ? 'bg-card text-card-foreground shadow-sm'
            : 'text-muted-foreground hover:text-foreground'"
          @click="update('protocolType', p.value)"
        >
          {{ p.label }}
        </button>
      </div>
    </div>

    <div class="border-t border-border" />

    <!-- Address & Port -->
    <div class="grid grid-cols-[1fr_100px] gap-2">
      <FormField label="Address">
        <input
          :value="model.address"
          type="text"
          placeholder="example.com"
          class="field-input"
          @input="update('address', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
      <FormField label="Port">
        <input
          :value="model.port"
          type="number"
          placeholder="443"
          class="field-input"
          @input="update('port', parseInt(($event.target as HTMLInputElement).value) || 443)"
        />
      </FormField>
    </div>

    <div class="border-t border-border" />

    <!-- Protocol-specific fields -->
    <template v-if="model.protocolType === 'vmess'">
      <FormField label="UUID">
        <input
          :value="model.uuid"
          type="text"
          placeholder="uuid"
          class="field-input"
          @input="update('uuid', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
      <FormField label="Alter ID">
        <input
          :value="model.alterId"
          type="number"
          placeholder="0"
          class="field-input"
          @input="update('alterId', parseInt(($event.target as HTMLInputElement).value) || 0)"
        />
      </FormField>
    </template>

    <template v-if="model.protocolType === 'vless'">
      <FormField label="UUID">
        <input
          :value="model.uuid"
          type="text"
          placeholder="uuid"
          class="field-input"
          @input="update('uuid', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
      <FormField label="Flow">
        <input
          :value="model.flow"
          type="text"
          placeholder="xtls-rprx-vision"
          class="field-input"
          @input="update('flow', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
    </template>

    <template v-if="model.protocolType === 'trojan'">
      <FormField label="Password">
        <input
          :value="model.password"
          type="password"
          placeholder="password"
          class="field-input"
          @input="update('password', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
    </template>

    <template v-if="model.protocolType === 'ss'">
      <div class="space-y-1.5">
        <label class="text-xs font-medium text-muted-foreground">Cipher</label>
        <select
          :value="model.cipher"
          class="field-input"
          @change="update('cipher', ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="c in ciphers" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <FormField label="Password">
        <input
          :value="model.password"
          type="password"
          placeholder="password"
          class="field-input"
          @input="update('password', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
    </template>

    <div class="border-t border-border" />

    <!-- Transport -->
    <div class="space-y-1.5">
      <label class="text-xs font-medium text-muted-foreground">Transport</label>
      <div class="grid grid-cols-6 gap-1 p-0.5 rounded-lg bg-muted">
        <button
          v-for="t in transports"
          :key="t.value"
          class="h-8 rounded-md text-xs font-medium transition-all duration-150"
          :class="model.transport === t.value
            ? 'bg-card text-card-foreground shadow-sm'
            : 'text-muted-foreground hover:text-foreground'"
          @click="update('transport', t.value)"
        >
          {{ t.label }}
        </button>
      </div>
    </div>

    <!-- TLS -->
    <div class="flex items-center justify-between py-1">
      <span class="text-sm text-foreground">TLS</span>
      <button
        class="relative w-10 h-[22px] rounded-full transition-colors duration-150"
        :class="model.tls ? 'bg-blue-600' : 'bg-muted'"
        role="switch"
        :aria-checked="model.tls"
        @click="update('tls', !model.tls)"
      >
        <span
          class="absolute top-[3px] left-[3px] w-4 h-4 rounded-full bg-white shadow-sm transition-transform duration-150"
          :class="{ 'translate-x-[18px]': model.tls }"
        />
      </button>
    </div>

    <!-- Transport-specific fields -->
    <template v-if="['ws', 'h2', 'httpupgrade', 'splithttp', 'xhttp'].includes(model.transport)">
      <FormField label="Host">
        <input
          :value="model.host"
          type="text"
          placeholder="example.com"
          class="field-input"
          @input="update('host', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
      <FormField label="Path">
        <input
          :value="model.path"
          type="text"
          placeholder="/"
          class="field-input"
          @input="update('path', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
    </template>

    <template v-if="model.transport === 'grpc'">
      <FormField label="Service Name">
        <input
          :value="model.serviceName"
          type="text"
          placeholder="grpcService"
          class="field-input"
          @input="update('serviceName', ($event.target as HTMLInputElement).value)"
        />
      </FormField>
    </template>

    <!-- SNI -->
    <FormField label="SNI">
      <input
        :value="model.sni"
        type="text"
        placeholder="example.com"
        class="field-input"
        @input="update('sni', ($event.target as HTMLInputElement).value)"
      />
    </FormField>

    <!-- Fingerprint -->
    <div class="space-y-1.5">
      <label class="text-xs font-medium text-muted-foreground">Fingerprint</label>
      <select
        :value="model.fingerprint"
        class="field-input"
        @change="update('fingerprint', ($event.target as HTMLSelectElement).value)"
      >
        <option v-for="f in fingerprints" :key="f" :value="f">{{ f }}</option>
      </select>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Node, NodeProtocol, NodeTransport } from '@/types'
import FormField from './FormField.vue'

const props = defineProps<{
  model: Node
}>()

const emit = defineEmits<{
  'update:model': [node: Node]
}>()

const protocols: { value: NodeProtocol; label: string }[] = [
  { value: 'vmess', label: 'VMess' },
  { value: 'vless', label: 'VLESS' },
  { value: 'trojan', label: 'Trojan' },
  { value: 'ss', label: 'SS' },
]

const transports: { value: NodeTransport; label: string }[] = [
  { value: 'tcp', label: 'TCP' },
  { value: 'ws', label: 'WS' },
  { value: 'grpc', label: 'gRPC' },
  { value: 'h2', label: 'H2' },
  { value: 'httpupgrade', label: 'HUP' },
  { value: 'splithttp', label: 'SHT' },
  { value: 'xhttp', label: 'XHTTP' },
]

const ciphers = ['aes-128-gcm', 'aes-256-gcm', 'chacha20-poly1305', 'none']
const fingerprints = ['chrome', 'firefox', 'safari', 'edge', 'ios', 'android', 'random', 'randomized', 'none']

function update(key: string, value: any) {
  emit('update:model', { ...props.model, [key]: value })
}
</script>

<style scoped>
.field-input {
  @apply h-9 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150;
}
</style>
