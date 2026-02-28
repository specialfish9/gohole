import { ThemeToggle } from "@/components/theme-toggle"
import { Search } from "lucide-react"

export default () => {
  const toSearch = () => {
    window.location.href = "/search"
  }

  return <header className="border-b bg-card" >
    <div className="container mx-auto px-4 py-4">
      <div className="flex items-center justify-between">
        <a href="/" className="flex items-center space-x-4">
          <div className="flex items-center space-x-4">
            <img src="/gohole.png" alt="Gohole Logo" className="h-12 w-12" />
            <h1 className="text-2xl font-bold">Gohole</h1>
          </div>
        </a>
        <div className="flex items-center space-x-2">
          <button className="btn font-bold flex items-center gap-2 bg-secondary rounded-xl px-4 py-2" onClick={toSearch}>
            <Search className="h-5 w-5 text-xl" />
            Search
          </button>
          <ThemeToggle />
        </div>
      </div>
    </div>
  </header >
}
