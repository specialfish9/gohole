import { TimePicker } from "@/components/time-picker";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { DomainDetail, goholeAPI } from "@/lib/api";
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { ResponsiveContainer, LineChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Line, Legend } from "recharts"

const AllowedGranularities = [
  { value: "1m", label: "1 Minute" },
  { value: "1h", label: "1 Hour" },
  { value: "1d", label: "1 Day" }
]

export default function Domain() {
  const [searchParams] = useSearchParams();
  const d = searchParams.get("d");
  const [details, setDetails] = useState<DomainDetail>(null);
  const [timeInterval, setTimeInterval] = useState("24h")
  const [granularity, setGranularity] = useState("1h")
  const { toast } = useToast()

  const fetchDetails = async () => {
    try {
      const detail = await goholeAPI.getDomainDetails(d, timeInterval, granularity)
      setDetails(detail)
    } catch (error) {
      console.error('Failed to fetch domain details:', error)
      toast({
        title: "Error",
        description: "Failed to fetch query data. Check if the backend is running.",
        variant: "destructive",
      })
    }
  }

  useState(() => {
    if (!d) return;
    fetchDetails()
  })

  useEffect(() => {
    if (!d) return;
    fetchDetails()
  }, [timeInterval, granularity])

  const dateFormat = () => {
    if (granularity === "1m") {
      return {
        hour: "2-digit",
        minute: "2-digit"
      }
    } else if (granularity === "1h") {
      return {
        hour: "2-digit",
        minute: "2-digit"
      }
    } else if (granularity === "1d") {
      return {
        month: "numeric",
        day: "numeric"
      }
    }
    return {
      month: "numeric",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit"
    }
  }

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
                {label && new Date(label).toLocaleString("en-UK", dateFormat())}
              </span>
            </div>
            <div className="flex flex-col">
              <span className="text-[0.70rem] uppercase text-muted-foreground">
                Count
              </span>
              <span className="font-bold text-primary">
                {payload[0].payload.count}
              </span>
            </div>
          </div>
        </div>
      )
    }
    return null
  }

  return (
    <>
      {!d ? (
        <div className="p-4 flex flex-col items-center gap-20 justify-center">
          <h1 className="text-2xl font-bold mb-4">No domain specified</h1>
          <button onClick={() => window.location.href = "/"} className="btn background-transparent p-4 rounded-xl border border-primary text-primary hover:bg-primary/10">
            Go Back
          </button>
          <img src="/puzzled.png" alt="Gohole puzzled" className="h-40 w-40" />
        </div>
      ) : (
        <div className="p-4">
          <Card>
            <CardHeader>
              <CardTitle>
                <span className="text-muted-foreground">Stats for: </span>
                <span className="font-mono text-2xl">{d}</span>
                {details?.blocked ? (
                  <span className="ml-4 inline-flex items-center rounded-full border bg-destructive px-2.5 py-0.5 text-xs font-semibold text-destructive-foreground">
                    Blocked
                  </span>
                ) : (
                  <span className="ml-4 inline-flex items-center rounded-full border bg-green-500 px-2.5 py-0.5 text-xs font-semibold text-white">
                    Allowed
                  </span>
                )}
              </CardTitle>
              <TimePicker
                interval={timeInterval}
                onIntervalChange={setTimeInterval}
                granularity={granularity}
                onGranularityChange={setGranularity}
                granularityOptions={AllowedGranularities}
              />

            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart
                  style={{ width: '100%', aspectRatio: 1.618, margin: 'auto' }}
                  data={details?.points || []}>
                  <CartesianGrid stroke="#555" strokeDasharray="5 5" />
                  <XAxis
                    dataKey="time"
                    tickFormatter={
                      (value) => new Date(value)
                        .toLocaleString("en-UK", dateFormat())}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />
                  <YAxis width={100} />
                  <Line type="monotone" name={d + " queries"} dataKey="count" className="text-primary-foreground" />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </div>
      )}
    </>
  )
}
