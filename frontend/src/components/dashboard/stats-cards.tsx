import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Shield, Activity, AlertTriangle, CheckCircle } from "lucide-react"

interface StatsCardsProps {
  data: {
    totalQueries: number
    blockedQueries: number
    allowedQueries: number
    blockRate: number
  }
}

export function StatsCards({ data }: StatsCardsProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Queries</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{data.totalQueries.toLocaleString()}</div>
          <p className="text-xs text-muted-foreground">DNS requests processed</p>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Blocked Queries</CardTitle>
          <Shield className="h-4 w-4 text-destructive" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-destructive">{data.blockedQueries.toLocaleString()}</div>
          <p className="text-xs text-muted-foreground">Ads & trackers blocked</p>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Allowed Queries</CardTitle>
          <CheckCircle className="h-4 w-4 text-success" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-success">{data.allowedQueries.toLocaleString()}</div>
          <p className="text-xs text-muted-foreground">Legitimate requests</p>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Block Rate</CardTitle>
          <AlertTriangle className="h-4 w-4 text-warning" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{data.blockRate.toFixed(1)}%</div>
          <p className="text-xs text-muted-foreground">Queries blocked</p>
        </CardContent>
      </Card>
    </div>
  )
}