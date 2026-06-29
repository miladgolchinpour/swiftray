import type { Node, NodeProtocol, NodeTransport } from '@/types'

function urlEncode(s: string): string {
  return encodeURIComponent(s)
}

function encodeBase64(s: string): string {
  return btoa(s)
}

function encodeBase64NoPadding(s: string): string {
  return btoa(s).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')
}

function securityOrAuto(node: Node): string {
  return node.security || 'auto'
}

function toVMessURL(node: Node): string {
  const tlsStr = node.tls ? 'tls' : 'insecure'
  const json = {
    v: '2',
    ps: node.name,
    add: node.address,
    port: String(node.port),
    id: node.uuid,
    aid: String(node.alterId),
    scy: securityOrAuto(node),
    net: node.transport,
    type: 'none',
    host: node.host,
    path: node.path,
    tls: tlsStr,
    sni: node.sni,
    fp: node.fingerprint,
  }
  return 'vmess://' + encodeBase64(JSON.stringify(json))
}

function toVLessURL(node: Node): string {
  const params: string[] = [`type=${node.transport}`]
  if (node.tls) params.push('security=tls')
  if (node.sni) params.push('sni=' + urlEncode(node.sni))
  if (node.fingerprint) params.push('fp=' + urlEncode(node.fingerprint))
  if (node.host) params.push('host=' + urlEncode(node.host))
  if (node.path) params.push('path=' + urlEncode(node.path))
  if (node.flow) params.push('flow=' + urlEncode(node.flow))
  const query = params.join('&')
  const name = urlEncode(node.name)
  return `vless://${node.uuid}@${node.address}:${node.port}?${query}#${name}`
}

function toTrojanURL(node: Node): string {
  const params: string[] = [`type=${node.transport}`]
  if (node.tls) params.push('security=tls')
  if (node.sni) params.push('sni=' + urlEncode(node.sni))
  if (node.fingerprint) params.push('fp=' + urlEncode(node.fingerprint))
  if (node.host) params.push('host=' + urlEncode(node.host))
  if (node.path) params.push('path=' + urlEncode(node.path))
  const query = params.join('&')
  const name = urlEncode(node.name)
  return `trojan://${node.password}@${node.address}:${node.port}?${query}#${name}`
}

function toShadowsocksURL(node: Node): string {
  const cred = `${node.cipher}:${node.password}`
  const b64 = encodeBase64NoPadding(cred)
  const name = urlEncode(node.name)
  return `ss://${b64}@${node.address}:${node.port}#${name}`
}

export function configURL(node: Node): string {
  switch (node.protocolType) {
    case 'vmess':
      return toVMessURL(node)
    case 'vless':
      return toVLessURL(node)
    case 'trojan':
      return toTrojanURL(node)
    case 'ss':
      return toShadowsocksURL(node)
    default:
      return ''
  }
}

export function protocolDisplayName(p: NodeProtocol): string {
  return p.toUpperCase()
}

export function transportDisplayName(t: NodeTransport): string {
  return t
}

export function delayColorClass(delay: number | null): string {
  if (delay === null) return 'text-muted-foreground'
  if (delay === -1) return 'text-red-500'
  if (delay <= 200) return 'text-emerald-500'
  if (delay <= 1000) return 'text-amber-500'
  return 'text-orange-500'
}

export function delayString(delay: number | null): string {
  if (delay === null) return '-- ms'
  if (delay === -1) return 'Fail'
  return `${delay} ms`
}

export function configJSON(node: Node): string {
  const proto = node.protocolType === 'ss' ? 'shadowsocks' : node.protocolType

  const settings: any = {}
  if (proto === 'vmess' || proto === 'vless') {
    settings.vnext = [{
      address: node.address,
      port: node.port,
      users: [{
        id: node.uuid,
        ...(proto === 'vmess' ? { alterId: node.alterId, security: securityOrAuto(node) } : {}),
        ...(proto === 'vless' && node.flow ? { flow: node.flow } : {}),
      }],
    }]
  } else if (proto === 'trojan') {
    settings.servers = [{
      address: node.address,
      port: node.port,
      password: node.password,
    }]
  } else if (proto === 'shadowsocks') {
    settings.servers = [{
      address: node.address,
      port: node.port,
      method: node.cipher || 'aes-256-gcm',
      password: node.password,
    }]
  }

  const streamSettings: any = { network: node.transport }

  if (node.tls) {
    const security = node.realityPublicKey ? 'reality' : 'tls'
    streamSettings.security = security
    if (security === 'reality') {
      const rs: any = {}
      if (node.sni) rs.serverName = node.sni
      if (node.fingerprint) rs.fingerprint = node.fingerprint
      if (node.realityPublicKey) rs.publicKey = node.realityPublicKey
      if (node.realityShortId) rs.shortId = node.realityShortId
      if (node.realitySpiderX) rs.spiderX = node.realitySpiderX
      streamSettings.realitySettings = rs
    } else {
      const ts: any = {}
      if (node.sni) ts.serverName = node.sni
      if (node.fingerprint) ts.fingerprint = node.fingerprint
      if (node.alpn) ts.alpn = node.alpn.split(',').map(s => s.trim()).filter(Boolean)
      streamSettings.tlsSettings = ts
    }
  }

  if (node.transport === 'ws') {
    const ws: any = {}
    if (node.path) ws.path = node.path
    if (node.host) ws.headers = { Host: node.host }
    streamSettings.wsSettings = ws
  } else if (node.transport === 'grpc') {
    if (node.serviceName) streamSettings.grpcSettings = { serviceName: node.serviceName }
  } else if (['h2', 'httpupgrade', 'splithttp'].includes(node.transport)) {
    const hs: any = {}
    if (node.path) hs.path = node.path
    if (node.host) hs.host = [node.host]
    streamSettings.httpSettings = hs
  }

  const outbound = { protocol: proto, settings, streamSettings }
  return JSON.stringify(outbound, null, 2)
}
