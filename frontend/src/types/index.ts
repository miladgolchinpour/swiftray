export type NodeProtocol = 'vmess' | 'vless' | 'trojan' | 'ss'
export type NodeTransport = 'tcp' | 'ws' | 'grpc' | 'httpupgrade' | 'splithttp' | 'h2'

export interface Node {
  id: string
  name: string
  address: string
  port: number
  protocolType: NodeProtocol
  transport: NodeTransport
  tls: boolean
  delay: number | null
  uuid: string
  password: string
  alterId: number
  security: string
  cipher: string
  flow: string
  encryption: string
  sni: string
  fingerprint: string
  alpn: string
  realityPublicKey: string
  realityShortId: string
  realitySpiderX: string
  host: string
  path: string
  serviceName: string
  serviceMode: string
  rawLink: string
}

export interface Subscription {
  id: string
  name: string
  url: string
  nodes: Node[]
  lastUpdated: string | null
}

export interface AppSettings {
  httpPort: number
  socksPort: number
  mixedPort: boolean
  enableUDP: boolean
  allowLAN: boolean
  routeOnly: boolean
  enableSniffing: boolean
  sniffHTTP: boolean
  sniffTLS: boolean
  sniffQUIC: boolean
  sniffFakeDNS: boolean
  useProxyAuth: boolean
  proxyUsername: string
  proxyPassword: string
  defaultFingerprint: string
  enableFragment: boolean
  fragmentPackLength: string
  fragmentSleep: string
  fragmentInterval: string
  localDNS: string
  remoteDNS: string
  bootstrapDNS: string
  parallelQuery: boolean
  serveStale: boolean
  useSystemHosts: boolean
  customDNSHosts: boolean
  fakeIP: boolean
  blockSVCBHTTPS: boolean
  validateRegionalDomain: number
  enableSystemProxy: boolean
  enableMenuBar: boolean
  routingMode: number
  domainStrategy: number
  bypassIran: boolean
  bypassRussia: boolean
  bypassChina: boolean
  pingTestURL: string
  customGeoSources: string
  exclusions: string
  urlTestTimeout: number
  urlTestConcurrency: number
}

export interface IPInfo {
  ipVersion: number | null
  ipAddress: string
  latitude: number
  longitude: number
  countryName: string
  countryCode: string
  capital: string
  phoneCodes: number[]
  timeZones: string[]
  zipCode: string
  cityName: string
  regionName: string
  regionCode: string
  continent: string
  continentCode: string
  currencies: string[]
  languages: string[]
  asn: string
  asnOrganization: string
  isProxy: boolean
}

export function countryFlag(code: string): string {
  if (!code || code.length !== 2) return ''
  const upper = code.toUpperCase()
  const r1 = upper.charCodeAt(0) - 65 + 0x1F1E6
  const r2 = upper.charCodeAt(1) - 65 + 0x1F1E6
  return String.fromCodePoint(r1, r2)
}

export type LogLevel = 'info' | 'debug' | 'warning' | 'error'

export interface LogEntry {
  id: string
  level: LogLevel
  message: string
  timestamp: string
}
