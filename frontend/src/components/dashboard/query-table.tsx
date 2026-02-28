import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Shield, CheckCircle, Search } from "lucide-react"
import { useEffect, useState } from "react"
import { goholeAPI, type Query } from "@/lib/api"
import { useToast } from "@/hooks/use-toast"
import { useNavigate, useSearchParams } from "react-router-dom"

export const QueryTable = () => {
  const [filter, setFilter] = useState("");
  const [queries, setQueries] = useState<Query[]>([])
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { toast } = useToast()

  const fetchQueries = async () => {
    try {
      const f = searchParams.get("q");
      const queryData = await goholeAPI.getQueries(f)
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

  // Auto-refresh queries every 30 seconds
  useEffect(() => {
    fetchQueries()
    const refreshInterval = setInterval(fetchQueries, 30000)
    return () => clearInterval(refreshInterval)
  }, [])

  useEffect(() => {
    fetchQueries()
  }, [searchParams])

  const handleSearch = () => {
    navigate(`/search?q=${encodeURIComponent(filter)}`);
  }

  const handleKeyDown = (e) => {
    if (e.key === "Enter") handleSearch();
  }

  return (
    <Card>
      <CardHeader className="flex justify-between items-start gap-4 sm:flex-row sm:items-center">
        <CardTitle>Recent DNS Queries</CardTitle>
        <div>
          <Badge variant="outline" className="flex p-4 rounded-lg">
            <input
              type="text"
              className="input bg-secondary inline-flex items-center rounded-full border p-2 text-xs font-semibold transition-colors focus:outline-none"
              placeholder="Filter by domain..."
              value={filter}
              onKeyUp={handleKeyDown}
              onChange={(e) => setFilter(e.target.value)}
            />
            <button onClick={handleSearch} className="ml-2">
              <Search className="h-5 w-5 text-xl ml-4" />
            </button>
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Domain</TableHead>
              <TableHead>ms</TableHead>
              <TableHead>Host</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Time</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {queries.map((query, index) => (
              <TableRow key={index}>
                <TableCell className="font-mono text-sm max-w-[300px] truncate">
                  <a href={"/domain?d=" + query.name} className="hover:underline">
                    {query.name}
                  </a>
                </TableCell>
                <TableCell className="font-mono text-sm max-w-[300px] truncate">
                  {query.millis}
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className="font-mono">
                    {query.host}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge variant="outline" className="font-mono">
                    {query.type}
                  </Badge>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    {query.blocked ? (
                      <>
                        <Shield className="h-4 w-4 text-destructive" />
                        <Badge variant="destructive">Blocked</Badge>
                      </>
                    ) : (
                      <>
                        <CheckCircle className="h-4 w-4 text-success" />
                        <Badge className="bg-success text-success-foreground">Allowed</Badge>
                      </>
                    )}
                  </div>
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {new Date(query.timestamp).toLocaleDateString('en-UK', {
                    hour: "2-digit",
                    minute: "2-digit",
                    day: "2-digit",
                    month: "short",
                    year: "2-digit"
                  }) || new Date().toLocaleTimeString()}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
