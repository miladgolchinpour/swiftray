import * as GoApp from '../../wailsjs/go/main/App'
import type { models } from '../../wailsjs/go/models'

type AppSettings = models.AppSettings

interface APIResponse {
  ok: boolean
  error?: string
  data?: any
  message?: string
}

class APIClient {
  private async call<T>(fn: () => Promise<APIResponse>, context: string): Promise<{ data: T | null; error: string | null }> {
    try {
      const res = await fn()
      if (!res.ok) {
        const msg = res.error || 'Unknown error'
        console.error(`[API] ${context}: ${msg}`)
        return { data: null, error: msg }
      }
      console.log(`[API] ${context}: OK`)
      return { data: res.data as T, error: null }
    } catch (e: any) {
      const msg = e?.message || String(e)
      console.error(`[API] ${context} exception: ${msg}`)
      return { data: null, error: msg }
    }
  }

  private async callMsg(fn: () => Promise<APIResponse>, context: string): Promise<{ ok: boolean; error: string | null }> {
    try {
      const res = await fn()
      if (!res.ok) {
        const msg = res.error || 'Unknown error'
        console.error(`[API] ${context}: ${msg}`)
        return { ok: false, error: msg }
      }
      console.log(`[API] ${context}: OK - ${res.message || ''}`)
      return { ok: true, error: null }
    } catch (e: any) {
      const msg = e?.message || String(e)
      console.error(`[API] ${context} exception: ${msg}`)
      return { ok: false, error: msg }
    }
  }

  // System
  async getAppInfo() {
    return this.call<{ startupComplete: boolean; xrayInstalled: boolean; xrayVersion: string; geoIPReady: boolean; geoSiteReady: boolean; dataDir: string }>(GoApp.GetAppInfo, 'GetAppInfo')
  }
  async verifyReady() {
    return this.callMsg(GoApp.VerifyReady, 'VerifyReady')
  }
  async getResourceStatus() {
    return this.call<any>(GoApp.GetResourceStatus, 'GetResourceStatus')
  }

  // Settings
  async getSettings() {
    return this.call<AppSettings>(GoApp.GetSettings, 'GetSettings')
  }
  async saveSettings(settings: any) {
    return this.callMsg(() => GoApp.SaveSettings(settings), 'SaveSettings')
  }
  async saveAndReload(settings: any) {
    return this.callMsg(() => GoApp.SaveAndReload(settings), 'SaveAndReload')
  }

  // Local Nodes
  async getLocalNodes() {
    return this.call<any[]>(GoApp.GetLocalNodes, 'GetLocalNodes')
  }
  async saveLocalNodes(nodes: any[]) {
    return this.callMsg(() => GoApp.SaveLocalNodes(nodes), 'SaveLocalNodes')
  }
  async addLocalNode(node: any) {
    return this.call<any[]>(() => GoApp.AddLocalNode(node), 'AddLocalNode')
  }
  async updateLocalNode(node: any) {
    return this.call<any[]>(() => GoApp.UpdateLocalNode(node), 'UpdateLocalNode')
  }
  async deleteLocalNode(id: string) {
    return this.call<any[]>(() => GoApp.DeleteLocalNode(id), 'DeleteLocalNode')
  }
  async getSelectedNodeID() {
    return this.call<string>(GoApp.GetSelectedNodeID, 'GetSelectedNodeID')
  }
  async setSelectedNodeID(id: string) {
    return this.callMsg(() => GoApp.SetSelectedNodeID(id), 'SetSelectedNodeID')
  }

  // Subscriptions
  async getSubscriptions() {
    return this.call<any[]>(GoApp.GetSubscriptions, 'GetSubscriptions')
  }
  async addSubscription(name: string, url: string) {
    return this.call<any[]>(() => GoApp.AddSubscription(name, url), 'AddSubscription')
  }
  async updateSubscription(id: string, name: string, url: string) {
    return this.call<any[]>(() => GoApp.UpdateSubscription(id, name, url), 'UpdateSubscription')
  }
  async deleteSubscription(id: string) {
    return this.call<any[]>(() => GoApp.DeleteSubscription(id), 'DeleteSubscription')
  }
  async refreshSubscription(id: string) {
    return this.call<any[]>(() => GoApp.RefreshSubscription(id), 'RefreshSubscription')
  }
  async refreshAllSubscriptions() {
    return this.call<any[]>(GoApp.RefreshAllSubscriptions, 'RefreshAllSubscriptions')
  }
  async getSelectedSubID() {
    return this.call<string>(GoApp.GetSelectedSubID, 'GetSelectedSubID')
  }
  async setSelectedSubID(id: string) {
    return this.callMsg(() => GoApp.SetSelectedSubID(id), 'SetSelectedSubID')
  }

  // Connection
  async connect() {
    return this.callMsg(GoApp.Connect, 'Connect')
  }
  async disconnect() {
    return this.callMsg(GoApp.Disconnect, 'Disconnect')
  }
  async getConnectionState() {
    return this.call<{ state: string; error?: string; message?: string }>(GoApp.GetConnectionState, 'GetConnectionState')
  }

  // Updates
  async getXrayVersion() {
    return this.call<string>(GoApp.GetXrayVersion, 'GetXrayVersion')
  }
  async checkXrayUpdate() {
    return this.call<boolean>(GoApp.CheckXrayUpdate, 'CheckXrayUpdate')
  }

  // Runtime Updater
  async checkXrayUpdateStatus() {
    return this.call<any>(GoApp.CheckXrayUpdateStatus, 'CheckXrayUpdateStatus')
  }
  async downloadXrayUpdate() {
    return this.callMsg(GoApp.DownloadXrayUpdate, 'DownloadXrayUpdate')
  }
  async cancelDownload() {
    return this.callMsg(GoApp.CancelDownload, 'CancelDownload')
  }

  // Logs
  async getLogs() {
    return this.call<any[]>(GoApp.GetLogs, 'GetLogs')
  }
  async clearLogs() {
    return this.callMsg(GoApp.ClearLogs, 'ClearLogs')
  }

  // IP Info
  async getIPInfo() {
    return this.call<any>(GoApp.GetIPInfo, 'GetIPInfo')
  }
  async fetchIPInfo() {
    return this.call<any>(GoApp.FetchIPInfo, 'FetchIPInfo')
  }

  // URL Test
  async urlTest() {
    return this.callMsg(GoApp.URLTest, 'URLTest')
  }
  async urlTestLocal() {
    return this.callMsg(GoApp.URLTestLocal, 'URLTestLocal')
  }
  async urlTestNodes(nodes: any[]) {
    return this.callMsg(() => GoApp.URLTestNodes(nodes), 'URLTestNodes')
  }
}

export const api = new APIClient()
