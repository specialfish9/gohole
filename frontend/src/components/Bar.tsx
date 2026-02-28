import { ThemeToggle } from "@/components/theme-toggle"

export default () => {
  const toSearch = () => {
    window.location.href = "/search"
  }

  return <header className="border-b bg-card" >
    <div className="container mx-auto px-4 py-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <img src="/gohole.png" alt="Gohole Logo" className="h-12 w-12" />
          <h1 className="text-2xl font-bold">Gohole</h1>
        </div>
        <div className="flex items-center space-x-2">
          <button className="btn btn-sm font-bold" onClick={toSearch}>Search</button>
          <ThemeToggle />
        </div>
      </div>
    </div>
  </header >
}
