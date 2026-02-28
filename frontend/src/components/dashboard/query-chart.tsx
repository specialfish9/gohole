import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Label } from "@/components/ui/label"
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip } from "recharts"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Shield, CheckCircle, Activity, AlertTriangle } from "lucide-react"
import { DomainStats, HostStat } from "@/lib/api"

interface QueryChartProps {
  pieData: Array<{ name: string; value: number; color: string }>
  hostData: Array<HostStat>
  barData: Array<{ time: string; blocked: number; allowed: number }>
  domainStats?: DomainStats,
  interval: string
  granularity: string
  onIntervalChange: (interval: string) => void
  onGranularityChange: (granularity: string) => void
}

const intervalOptions = [
  { value: "1h", label: "Last Hour" },
  { value: "6h", label: "Last 6 Hours" },
  { value: "24h", label: "Last 24 Hours" },
  { value: "7d", label: "Last 7 Days" },
  { value: "30d", label: "Last 30 Days" }
]

const granularityOptions = [
  { value: "1m", label: "1 Minute" },
  { value: "5m", label: "5 Minutes" },
  { value: "15m", label: "15 Minutes" },
  { value: "1h", label: "1 Hour" },
  { value: "6h", label: "6 Hours" },
  { value: "1d", label: "1 Day" }
]

const hostPieColors = [
  "#8884d8", "#82ca9d", "#ffc658", "#ff8042", "#8dd1e1",
  "#a4de6c", "#d0ed57", "#ffc0cb", "#b0e0e6", "#f4a460"
];

export function QueryChart({
  pieData,
  hostData,
  barData,
  domainStats,
  interval,
  granularity,
  onIntervalChange,
  onGranularityChange
}: QueryChartProps) {
  const hostPieData = hostData.map((host) => ({
    name: host.host,
    value: host.queryCount,
  }))

  hostData.sort((a, b) => b.queryCount - a.queryCount)

  // Custom tooltip for bar chart
  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="rounded-lg border bg-background p-2 shadow-sm">
          <div className="grid grid-cols-2 gap-2">
            <div className="flex flex-col">
              <span className="text-[0.70rem] uppercase text-muted-foreground">
                Time
              </span>
              <span className="font-bold text-muted-foreground">
                {label}
              </span>
            </div>
            <div className="flex flex-col">
              <span className="text-[0.70rem] uppercase text-muted-foreground">
                Blocked
              </span>
              <span className="font-bold text-destructive">
                {payload[0].value}
              </span>
            </div>
            <div className="flex flex-col">
              <span className="text-[0.70rem] uppercase text-muted-foreground">
                Allowed
              </span>
              <span className="font-bold text-success">
                {payload[1].value}
              </span>
            </div>
          </div>
        </div>
      )
    }
    return null
  }

  return (
    <div>
      <div className="grid grid-cols-4 gap-4 mb-10">
        <div className="space-y-2">
          <Label htmlFor="interval-select">Time Interval</Label>
          <Select value={interval} onValueChange={onIntervalChange}>
            <SelectTrigger id="interval-select">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {intervalOptions.map(option => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="granularity-select">Granularity</Label>
          <Select value={granularity} onValueChange={onGranularityChange}>
            <SelectTrigger id="granularity-select">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {granularityOptions.map(option => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Query Distribution</CardTitle>
            <div className="text-sm text-muted-foreground">
              Current period: {intervalOptions.find(opt => opt.value === interval)?.label}
            </div>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={120}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Queries Over Time</CardTitle>
            <div className="text-sm text-muted-foreground">
              Current period: {intervalOptions.find(opt => opt.value === interval)?.label}
            </div>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={barData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="time"
                  className="text-xs fill-muted-foreground"
                  tick={{ fontSize: 12 }}
                />
                <YAxis
                  className="text-xs fill-muted-foreground"
                  tick={{ fontSize: 12 }}
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Bar
                  dataKey="blocked"
                  name="Blocked"
                  fill="hsl(var(--destructive))"
                  radius={[2, 2, 0, 0]}
                />
                <Bar
                  dataKey="allowed"
                  name="Allowed"
                  fill="hsl(var(--success))"
                  radius={[2, 2, 0, 0]}
                />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="col-span-2">
          <CardHeader>
            <CardTitle>Domain stats</CardTitle>
            <div className="text-sm text-muted-foreground">
              Current period: {intervalOptions.find(opt => opt.value === interval)?.label}
            </div>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" >
              <div>
                <div className="grid grid-cols-4 gap-4">
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">Queried domains</CardTitle>
                      <Activity className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{domainStats?.total ?? 0}</div>
                      <p className="text-xs text-muted-foreground">domains</p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">Blocked domains</CardTitle>
                      <Shield className="h-4 w-4 text-destructive" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold text-destructive">{domainStats?.blocked ?? 0}</div>
                      <p className="text-xs text-muted-foreground">domains</p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">Allowed domains</CardTitle>
                      <CheckCircle className="h-4 w-4 text-success" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold text-success">{domainStats ? domainStats.total - domainStats?.blocked : 0}</div>
                      <p className="text-xs text-muted-foreground">domains</p>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                      <CardTitle className="text-sm font-medium">Block Rate</CardTitle>
                      <AlertTriangle className="h-4 w-4 text-warning" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{domainStats
                        ? Math.round(domainStats.blocked * 100 / domainStats.total)
                        : 0}%</div>
                      <p className="text-xs text-muted-foreground">domains blocked</p>
                    </CardContent>
                  </Card>
                </div>
                <div className="grid grid-cols-2 gap-4 mt-8">
                  <div className="flex items-center gap-2">
                    <h3 className="text-xl font-bold font-medium">Top blocked domains</h3>
                    <Shield className="h-6 w-6 text-destructive" />
                  </div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-xl font-bold font-medium">Top allowed domains</h3>
                    <CheckCircle className="h-6 w-6 text-success" />
                  </div>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead></TableHead>
                        <TableHead>Domain</TableHead>
                        <TableHead>Hits</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {(domainStats?.topBlocked ?? []).map((topBlocked, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-mono text-sm max-w-[300px] truncate">
                            {index + 1}
                          </TableCell>
                          <TableCell className="font-mono text-sm max-w-[300px] truncate">
                            {topBlocked.domain}
                          </TableCell>
                          <TableCell>
                            <Badge variant="destructive">{topBlocked.count ?? 0}</Badge>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead></TableHead>
                        <TableHead>Domain</TableHead>
                        <TableHead>Hits</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {(domainStats?.topAllowed ?? []).map((topAllowed, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-mono text-sm max-w-[300px] truncate">
                            {index + 1}
                          </TableCell>
                          <TableCell className="font-mono text-sm max-w-[300px] truncate">
                            {topAllowed.domain}
                          </TableCell>
                          <TableCell>
                            <Badge className="bg-success text-success-foreground">{topAllowed.count ?? 0}</Badge>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </div>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="col-span-2">
          <CardHeader>
            <CardTitle>Hosts activity </CardTitle>
            <div className="text-sm text-muted-foreground">
              Current period: {intervalOptions.find(opt => opt.value === interval)?.label}
            </div>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-2">
            { /* PIE CHART */}
            <ResponsiveContainer width="100%" height={500}>
              <PieChart>
                <Pie
                  data={hostPieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={120}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {hostData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={hostPieColors[index]} />
                  ))}
                </Pie>
                <Legend />
              </PieChart>
            </ResponsiveContainer>

            { /* TABLE */}
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Host</TableHead>
                  <TableHead>Queries</TableHead>
                  <TableHead>Blocked</TableHead>
                  <TableHead>Block rate</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {hostData.map((hostStat, index) => (
                  <TableRow key={index}>
                    <TableCell className="font-mono text-sm max-w-[300px] truncate">
                      {hostStat.host}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="font-mono">
                        {hostStat.queryCount}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="font-mono">
                        {hostStat.blockedCount}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {true ? (
                          <>
                            <Shield className="h-4 w-4 text-destructive" />
                            <Badge variant="destructive">{hostStat.blockRate}%</Badge>
                          </>
                        ) : (
                          <>
                            <CheckCircle className="h-4 w-4 text-success" />
                            <Badge className="bg-success text-success-foreground">Allowed</Badge>
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div >
  )
}
