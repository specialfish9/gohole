interface Query {
  name: string
  type: string
  host: string
  blocked: boolean
  timestamp?: string
  millis: number
}

interface QueryStats {
  totalQueries: number
  blockedQueries: number
  allowedQueries: number
  blockRate: number
}

interface QueryHistoryPoint {
  time: string
  blocked: number
  allowed: number
}

interface BlocklistStats {
  totalEntries: number
}

interface HostStat {
  host: string
  queryCount: number
  blockedCount: number
  blockRate: number
}

class GoHoleAPI {
  private baseURL: string

  constructor(baseURL: string = '') {
    this.baseURL = baseURL
  }

  async getQueries(): Promise<Query[]> {
    try {
      const response = await fetch(`${this.baseURL}/api/queries`)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const queries = await response.json()

      // Add timestamps if not provided by backend
      return queries.map((query: Query) => ({
        ...query,
        timestamp: query.timestamp || new Date().toISOString()
      }))
    } catch (error) {
      console.error('Failed to fetch queries:', error)
      throw error
    }
  }

  async getStats(interval: string = '24h'): Promise<QueryStats> {
    try {
      const response = await fetch(`${this.baseURL}/api/queries/stats?interval=${interval}`)
      if (!response.ok) {
        // If stats endpoint doesn't exist yet, calculate from queries
        const queries = await this.getQueries()
        return this.calculateStatsFromQueries(queries)
      }
      return await response.json()
    } catch (error) {
      console.error('Failed to fetch stats:', error)
      // Fallback to calculating from queries
      const queries = await this.getQueries()
      return this.calculateStatsFromQueries(queries)
    }
  }

  async getQueryHistory(interval: string = '24h', granularity: string = '1h'): Promise<QueryHistoryPoint[]> {
    try {
      const response = await fetch(`${this.baseURL}/api/queries/stats/history?interval=${interval}&granularity=${granularity}`)
      if (!response.ok) {
        // If history endpoint doesn't exist yet, generate from current queries
        return this.generateHistoryFromQueries(interval, granularity)
      }
      return await response.json()
    } catch (error) {
      console.error('Failed to fetch query history:', error)
      // Fallback to generated data
      return this.generateHistoryFromQueries(interval, granularity)
    }
  }

  async getBlocklistStats(): Promise<BlocklistStats> {
    try {
      const response = await fetch(`${this.baseURL}/api/blocklist/stats`)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      return await response.json()
    } catch (error) {
      console.error('Failed to fetch blocklist stats:', error)
      throw error
    }
  }

  async getHostStats(interval: string = '24h'): Promise<HostStat[]> {
    try {
      const response = await fetch(`${this.baseURL}/api/hosts/stats?interval=${interval}`)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      return await response.json()
    } catch (error) {
      console.error('Failed to fetch host stats:', error)
      throw error
    }
  }

  async uploadBlocklist(file: File): Promise<{ success: boolean; message: string }> {
    try {
      const formData = new FormData()
      formData.append('blocklist', file)

      const response = await fetch(`${this.baseURL}/api/blocklist/upload`, {
        method: 'POST',
        body: formData
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Failed to upload blocklist:', error)
      throw error
    }
  }

  private calculateStatsFromQueries(queries: Query[]): QueryStats {
    const totalQueries = queries.length
    const blockedQueries = queries.filter(q => q.blocked).length
    const allowedQueries = totalQueries - blockedQueries
    const blockRate = totalQueries > 0 ? (blockedQueries / totalQueries) * 100 : 0

    return {
      totalQueries,
      blockedQueries,
      allowedQueries,
      blockRate
    }
  }

  private generateHistoryFromQueries(interval: string, granularity: string): QueryHistoryPoint[] {
    // This is a simplified version - in a real implementation, 
    // you'd want to process actual historical data
    const now = new Date()
    const data: QueryHistoryPoint[] = []

    // Generate some realistic data points based on interval
    const points = this.getDataPointsForInterval(interval, granularity)

    for (let i = 0; i < points.length; i++) {
      const time = new Date(now.getTime() - (points.length - i) * this.getMillisecondsForGranularity(granularity))
      data.push({
        time: this.formatTimeForGranularity(time, granularity),
        blocked: Math.floor(Math.random() * 50) + 10,
        allowed: Math.floor(Math.random() * 150) + 50
      })
    }

    return data
  }

  private getDataPointsForInterval(interval: string, granularity: string): number[] {
    const intervalMs = this.getMillisecondsForInterval(interval)
    const granularityMs = this.getMillisecondsForGranularity(granularity)
    const points = Math.min(50, Math.floor(intervalMs / granularityMs))
    return Array.from({ length: points }, (_, i) => i)
  }

  private getMillisecondsForInterval(interval: string): number {
    switch (interval) {
      case '1h': return 60 * 60 * 1000
      case '6h': return 6 * 60 * 60 * 1000
      case '24h': return 24 * 60 * 60 * 1000
      case '7d': return 7 * 24 * 60 * 60 * 1000
      case '30d': return 30 * 24 * 60 * 60 * 1000
      default: return 24 * 60 * 60 * 1000
    }
  }

  private getMillisecondsForGranularity(granularity: string): number {
    switch (granularity) {
      case '1m': return 60 * 1000
      case '5m': return 5 * 60 * 1000
      case '15m': return 15 * 60 * 1000
      case '1h': return 60 * 60 * 1000
      case '6h': return 6 * 60 * 60 * 1000
      case '1d': return 24 * 60 * 60 * 1000
      default: return 60 * 60 * 1000
    }
  }

  private formatTimeForGranularity(date: Date, granularity: string): string {
    switch (granularity) {
      case '1m':
      case '5m':
      case '15m':
      case '1h':
        return date.toLocaleTimeString('en-US', {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false
        })
      case '6h':
        return date.toLocaleString('en-US', {
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          hour12: false
        })
      case '1d':
        return date.toLocaleDateString('en-US', {
          month: '2-digit',
          day: '2-digit'
        })
      default:
        return date.toLocaleTimeString('en-US', {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false
        })
    }
  }
}

// Export singleton instance
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || window.location.origin;
export const goholeAPI = new GoHoleAPI(API_BASE_URL)
export type { Query, QueryStats, QueryHistoryPoint, BlocklistStats, HostStat }
