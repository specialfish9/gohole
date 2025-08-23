import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Label } from "@/components/ui/label"
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip } from "recharts"

interface QueryChartProps {
  pieData: Array<{ name: string; value: number; color: string }>
  barData: Array<{ time: string; blocked: number; allowed: number }>
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

export function QueryChart({ 
  pieData, 
  barData, 
  interval, 
  granularity, 
  onIntervalChange, 
  onGranularityChange 
}: QueryChartProps) {
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
          <div className="grid grid-cols-2 gap-4">
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
    </div>
  )
}