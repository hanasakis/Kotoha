export default function PriceDisplay({ price, className }: { price: number; className?: string }) {
  const yuan = (price / 100).toFixed(2)
  return <span className={className}>¥{yuan}</span>
}
