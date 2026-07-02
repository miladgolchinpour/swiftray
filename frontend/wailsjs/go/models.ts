export namespace main {
	
	export class APIResponse {
	    ok: boolean;
	    error?: string;
	    data?: any;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new APIResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.error = source["error"];
	        this.data = source["data"];
	        this.message = source["message"];
	    }
	}

}

export namespace models {
	
	export class AppSettings {
	    httpPort: number;
	    socksPort: number;
	    mixedPort: boolean;
	    enableUDP: boolean;
	    allowLAN: boolean;
	    routeOnly: boolean;
	    enableSniffing: boolean;
	    sniffHTTP: boolean;
	    sniffTLS: boolean;
	    sniffQUIC: boolean;
	    sniffFakeDNS: boolean;
	    useProxyAuth: boolean;
	    proxyUsername: string;
	    proxyPassword: string;
	    defaultFingerprint: string;
	    enableFragment: boolean;
	    fragmentPackLength: string;
	    fragmentSleep: string;
	    fragmentInterval: string;
	    localDNS: string;
	    remoteDNS: string;
	    bootstrapDNS: string;
	    parallelQuery: boolean;
	    serveStale: boolean;
	    useSystemHosts: boolean;
	    customDNSHosts: boolean;
	    fakeIP: boolean;
	    blockSVCBHTTPS: boolean;
	    validateRegionalDomain: number;
	    enableSystemProxy: boolean;
	    enableMenuBar: boolean;
	    routingMode: number;
	    domainStrategy: number;
	    bypassIran: boolean;
	    bypassRussia: boolean;
	    bypassChina: boolean;
	    pingTestURL: string;
	    customGeoSources: string;
	    exclusions: string;
	    urlTestMode: string;
	    urlTestTimeout: number;
	    urlTestConcurrency: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.httpPort = source["httpPort"];
	        this.socksPort = source["socksPort"];
	        this.mixedPort = source["mixedPort"];
	        this.enableUDP = source["enableUDP"];
	        this.allowLAN = source["allowLAN"];
	        this.routeOnly = source["routeOnly"];
	        this.enableSniffing = source["enableSniffing"];
	        this.sniffHTTP = source["sniffHTTP"];
	        this.sniffTLS = source["sniffTLS"];
	        this.sniffQUIC = source["sniffQUIC"];
	        this.sniffFakeDNS = source["sniffFakeDNS"];
	        this.useProxyAuth = source["useProxyAuth"];
	        this.proxyUsername = source["proxyUsername"];
	        this.proxyPassword = source["proxyPassword"];
	        this.defaultFingerprint = source["defaultFingerprint"];
	        this.enableFragment = source["enableFragment"];
	        this.fragmentPackLength = source["fragmentPackLength"];
	        this.fragmentSleep = source["fragmentSleep"];
	        this.fragmentInterval = source["fragmentInterval"];
	        this.localDNS = source["localDNS"];
	        this.remoteDNS = source["remoteDNS"];
	        this.bootstrapDNS = source["bootstrapDNS"];
	        this.parallelQuery = source["parallelQuery"];
	        this.serveStale = source["serveStale"];
	        this.useSystemHosts = source["useSystemHosts"];
	        this.customDNSHosts = source["customDNSHosts"];
	        this.fakeIP = source["fakeIP"];
	        this.blockSVCBHTTPS = source["blockSVCBHTTPS"];
	        this.validateRegionalDomain = source["validateRegionalDomain"];
	        this.enableSystemProxy = source["enableSystemProxy"];
	        this.enableMenuBar = source["enableMenuBar"];
	        this.routingMode = source["routingMode"];
	        this.domainStrategy = source["domainStrategy"];
	        this.bypassIran = source["bypassIran"];
	        this.bypassRussia = source["bypassRussia"];
	        this.bypassChina = source["bypassChina"];
	        this.pingTestURL = source["pingTestURL"];
	        this.customGeoSources = source["customGeoSources"];
	        this.exclusions = source["exclusions"];
	        this.urlTestMode = source["urlTestMode"];
	        this.urlTestTimeout = source["urlTestTimeout"];
	        this.urlTestConcurrency = source["urlTestConcurrency"];
	    }
	}
	export class Node {
	    id: string;
	    name: string;
	    address: string;
	    port: number;
	    protocolType: string;
	    transport: string;
	    tls: boolean;
	    delay?: number;
	    uuid: string;
	    password: string;
	    alterId: number;
	    security: string;
	    cipher: string;
	    flow: string;
	    encryption: string;
	    sni: string;
	    fingerprint: string;
	    alpn: string;
	    realityPublicKey: string;
	    realityShortId: string;
	    realitySpiderX: string;
	    host: string;
	    path: string;
	    serviceName: string;
	    serviceMode: string;
	    rawLink: string;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.port = source["port"];
	        this.protocolType = source["protocolType"];
	        this.transport = source["transport"];
	        this.tls = source["tls"];
	        this.delay = source["delay"];
	        this.uuid = source["uuid"];
	        this.password = source["password"];
	        this.alterId = source["alterId"];
	        this.security = source["security"];
	        this.cipher = source["cipher"];
	        this.flow = source["flow"];
	        this.encryption = source["encryption"];
	        this.sni = source["sni"];
	        this.fingerprint = source["fingerprint"];
	        this.alpn = source["alpn"];
	        this.realityPublicKey = source["realityPublicKey"];
	        this.realityShortId = source["realityShortId"];
	        this.realitySpiderX = source["realitySpiderX"];
	        this.host = source["host"];
	        this.path = source["path"];
	        this.serviceName = source["serviceName"];
	        this.serviceMode = source["serviceMode"];
	        this.rawLink = source["rawLink"];
	    }
	}

}

