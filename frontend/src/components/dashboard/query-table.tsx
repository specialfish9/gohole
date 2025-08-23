import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Shield, CheckCircle } from "lucide-react"

interface Query {
  name: string
  type: string
  blocked: boolean
  timestamp?: string
}

interface QueryTableProps {
  queries: Query[]
}

export function QueryTable({ queries }: QueryTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent DNS Queries</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Domain</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Time</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {queries.map((query, index) => (
              <TableRow key={index}>
                <TableCell className="font-mono text-sm max-w-[300px] truncate">
                  {query.name}
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
                  {query.timestamp || new Date().toLocaleTimeString()}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}