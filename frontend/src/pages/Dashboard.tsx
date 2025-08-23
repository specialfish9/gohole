import { useEffect, useState } from "react"
import { goholeAPI, type Query } from "@/lib/api"
import { StatsCards } from "@/components/dashboard/stats-cards"
import { QueryChart } from "@/components/dashboard/query-chart"
import { QueryTable } from "@/components/dashboard/query-table"
import { ThemeToggle } from "@/components/theme-toggle"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { RefreshCw, Settings, Upload } from "lucide-react"
import { useToast } from "@/hooks/use-toast"

export default function Dashboard() {
  const [queries, setQueries] = useState<Query[]>([])
  const [loading, setLoading] = useState(false)
  const [timeInterval, setTimeInterval] = useState("24h")
  const [granularity, setGranularity] = useState("1h")
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date())
  const { toast } = useToast()

  const fetchQueries = async () => {
    setLoading(true)
    try {
      const queryData = await goholeAPI.getQueries()
      setQueries(queryData)
      setLastUpdated(new Date())

      toast({
        title: "Queries updated",
        description: `Loaded ${queryData.length} DNS queries`,
      })
    } catch (error) {
      console.error('Failed to fetch queries:', error)
      toast({
        title: "Error",
        description: "Failed to fetch query data. Check if the backend is running.",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  // Auto-refresh queries every 30 seconds
  useEffect(() => {
    fetchQueries()

    const refreshInterval = setInterval(fetchQueries, 30000)
    return () => clearInterval(refreshInterval)
  }, [])

  // Fetch updated chart data when timeInterval or granularity changes
  useEffect(() => {
    fetchChartData()
  }, [timeInterval, granularity])

  const fetchChartData = async () => {
    try {
      // This will fall back to generated data if the backend endpoints don't exist yet
      await goholeAPI.getQueryHistory(timeInterval, granularity)
    } catch (error) {
      console.error('Failed to fetch chart data:', error)
    }
  }

  // Calculate statistics
  const totalQueries = queries.length
  const blockedQueries = queries.filter(q => q.blocked).length
  const allowedQueries = totalQueries - blockedQueries
  const blockRate = totalQueries > 0 ? (blockedQueries / totalQueries) * 100 : 0

  const statsData = {
    totalQueries,
    blockedQueries,
    allowedQueries,
    blockRate
  }

  // Chart data
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
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <img src="/gohole.png" alt="Gohole Logo" className="h-12 w-12" />
              <h1 className="text-2xl font-bold">Gohole</h1>
              <div className="text-xs text-muted-foreground">
                Last updated: {lastUpdated.toLocaleTimeString()}
              </div>
            </div>
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="icon"
                onClick={fetchQueries}
                disabled={loading}
              >
                <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              </Button>
              <Button variant="outline" size="icon">
                <Upload className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon">
                <Settings className="h-4 w-4" />
              </Button>
              <ThemeToggle />
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-6 space-y-6">
        {/* Statistics Cards */}
        <StatsCards data={statsData} />

        {/* Charts */}
        <QueryChart
          pieData={pieData}
          barData={barData}
          interval={timeInterval}
          granularity={granularity}
          onIntervalChange={setTimeInterval}
          onGranularityChange={setGranularity}
        />

        {/* Query Table */}
        <QueryTable queries={queries} />
      </main>
    </div>
  )
}
