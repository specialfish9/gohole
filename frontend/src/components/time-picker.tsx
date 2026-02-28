import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Label } from "@/components/ui/label"

interface IntervalOption {
  value: string;
  label: string;
}

interface GranularityOption {
  value: string;
  label: string;
}

interface TimePickerProps {
  interval: string
  granularity: string
  intervalOptions?: IntervalOption[]
  granularityOptions?: GranularityOption[]
  onIntervalChange: (interval: string) => void
  onGranularityChange: (granularity: string) => void
}


export const DefaultIntervalOptions: IntervalOption[] = [
  { value: "1h", label: "Last Hour" },
  { value: "6h", label: "Last 6 Hours" },
  { value: "24h", label: "Last 24 Hours" },
  { value: "7d", label: "Last 7 Days" },
  { value: "30d", label: "Last 30 Days" }
]

export const DefaultGranularityOptions: GranularityOption[] = [
  { value: "1m", label: "1 Minute" },
  { value: "5m", label: "5 Minutes" },
  { value: "15m", label: "15 Minutes" },
  { value: "1h", label: "1 Hour" },
  { value: "6h", label: "6 Hours" },
  { value: "1d", label: "1 Day" }
]
export const TimePicker = (opts: TimePickerProps) => {
  return <>
    <div className="grid grid-cols-4 gap-4 mb-10">
      <div className="space-y-2">
        <Label htmlFor="interval-select">Time Interval</Label>
        <Select value={opts.interval} onValueChange={opts.onIntervalChange}>
          <SelectTrigger id="interval-select">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {(opts.intervalOptions || DefaultIntervalOptions).map(option => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label htmlFor="granularity-select">Granularity</Label>
        <Select value={opts.granularity} onValueChange={opts.onGranularityChange}>
          <SelectTrigger id="granularity-select">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {(opts.granularityOptions || DefaultGranularityOptions).map(option => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </div>
  </>
}
