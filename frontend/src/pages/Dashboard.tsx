import { useEffect, useState } from "react"
import { goholeAPI, type Query, type BlocklistStats, type HostStat, QueryStats, DomainStats } from "@/lib/api"
import { StatsCards } from "@/components/dashboard/stats-cards"
import { QueryChart } from "@/components/dashboard/query-chart"
import { QueryTable } from "@/components/dashboard/query-table"
import { useToast } from "@/hooks/use-toast"

export default function Dashboard() {
  const [stats, setStats] = useState<QueryStats>()
  const [queries, setQueries] = useState<Query[]>([])
  const [timeInterval, setTimeInterval] = useState("24h")
  const [granularity, setGranularity] = useState("1h")
  const [blocklistStats, setBlocklistStats] = useState<BlocklistStats>()
  const [hostStats, setHostStats] = useState<HostStat[]>([])
  const [domainStats, setDomainStats] = useState<DomainStats>()
  const { toast } = useToast()

  const fetchQueries = async () => {
    try {
      const queryData = await goholeAPI.getQueries()
      setQueries(queryData)
    } catch (error) {
      console.error('Failed to fetch queries:', error)
      toast({
        title: "Error",
        description: "Failed to fetch query data. Check if the backend is running.",
        variant: "destructive",
      })
    }
  }

  const fetchStats = async () => {
    try {
      // This will fall back to calculating stats from queries if the endpoint doesn't exist yet
      const stats = await goholeAPI.getStats(timeInterval)
      setStats(stats)
    } catch (error) {
      console.error('Failed to fetch stats:', error)
    }
  }

  const fetchBlocklistStats = async () => {
    try {
      const blstats = await goholeAPI.getBlocklistStats()
      setBlocklistStats(blstats)
    } catch (error) {
      console.error('Failed to fetch blocklist stats:', error)
    }
  }

  const fetchHostStats = async () => {
    try {
      const hstats = await goholeAPI.getHostStats(timeInterval)
      setHostStats(hstats)
    } catch (error) {
      console.error('Failed to fetch host stats:', error)
    }
  }

  const fetchDomainStats = async () => {
    try {
      const dstats = await goholeAPI.getDomainStats(timeInterval)
      setDomainStats(dstats)
    } catch (error) {
      console.error('Failed to fetch domain stats:', error)
    }
  }

  // Auto-refresh queries every 30 seconds
  useEffect(() => {
    fetchStats()
    fetchQueries()
    fetchBlocklistStats()
    fetchHostStats()
    fetchDomainStats()

    const refreshInterval = setInterval(fetchQueries, 30000)
    return () => clearInterval(refreshInterval)
  }, [])

  // Fetch updated chart data when timeInterval or granularity changes
  useEffect(() => {
    fetchChartData()
    fetchHostStats()
    fetchStats()
  }, [timeInterval, granularity])

  const fetchChartData = async () => {
    try {
      // This will fall back to generated data if the backend endpoints don't exist yet
      await goholeAPI.getQueryHistory(timeInterval, granularity)
    } catch (error) {
      console.error('Failed to fetch chart data:', error)
    }
  }

  // Chart data
  const queriesInInterval = queries.length
  const blockedQueries = queries.filter(q => q.blocked).length
  const allowedQueries = queriesInInterval - blockedQueries
  const pieData = [
    { name: "Blocked", value: blockedQueries, color: "hsl(var(--destructive))" },
    { name: "Allowed", value: allowedQueries, color: "hsl(var(--success))" }
  ]

  // Generate chart data - this will be replaced by API calls when backend supports it
  const generateBarData = async () => {
    try {
      return await goholeAPI.getQueryHistory(timeInterval, granularity)
    } catch (error) {
      console.error('Failed to get query history:', error)
      // Fallback to empty data
      return []
    }
  }

  const [barData, setBarData] = useState<Array<{ time: string, blocked: number, allowed: number }>>([])

  useEffect(() => {
    const loadBarData = async () => {
      const data = await generateBarData()
      setBarData(data)
    }
    loadBarData()
  }, [timeInterval, granularity])

  return (
    <div className="min-h-screen bg-background">
      {/* Main Content */}
      <main className="container mx-auto px-4 py-6 space-y-6">
        {/* Statistics Cards */}
        <StatsCards data={{
          totalQueries: stats?.totalQueries || 0,
          blockedQueries: stats?.blockedQueries || 0,
          allowedQueries: stats?.allowedQueries || 0,
          blockRate: stats?.blockRate || 0,
          totalEntries: blocklistStats?.totalEntries || 0
        }} />

        {/* Charts */}
        <QueryChart
          pieData={pieData}
          barData={barData}
          hostData={hostStats}
          interval={timeInterval}
          granularity={granularity}
          domainStats={domainStats}
          onIntervalChange={setTimeInterval}
          onGranularityChange={setGranularity}
        />

      </main>
    </div>
  )
}
