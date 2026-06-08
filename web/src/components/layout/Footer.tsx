export default function Footer() {
  return (
    <footer className="bg-white border-t py-8 mt-12">
      <div className="max-w-7xl mx-auto px-4 text-center text-sm text-gray-400">
        <p className="mb-1">Kotoha — {new Date().getFullYear()}</p>
        <p>Powered by Go + Gin + React</p>
      </div>
    </footer>
  )
}
