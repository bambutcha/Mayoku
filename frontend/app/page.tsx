import Link from 'next/link'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'

export default function Home() {
  return (
    <>
      <Header />
      <main className="container mx-auto px-4 py-12">
        <div className="flex min-h-[calc(100vh-8rem)] flex-col items-center justify-center">
          {/* Hero Section */}
          <div className="text-center space-y-8 max-w-3xl">
            <div className="space-y-4">
              <h1 className="text-7xl md:text-8xl font-extrabold gradient-text pulse-glow">
                🕵️
              </h1>
              <h2 className="text-5xl md:text-6xl font-bold tracking-tight">
                <span className="gradient-text">Mayoku</span>
              </h2>
              <p className="text-xl md:text-2xl text-muted-foreground max-w-2xl mx-auto">
                Современная игра "Шпион" с новым уровнем интриги
              </p>
            </div>
            
            <div className="flex flex-col sm:flex-row gap-4 justify-center pt-4">
              <Link href="/lobby">
                <Button size="lg" className="hover-lift glow">
                  Найти игру
                </Button>
              </Link>
              <Link href="/profile">
                <Button variant="secondary" size="lg" className="hover-lift">
                  Профиль
                </Button>
              </Link>
            </div>
          </div>

          {/* Features Grid */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-20 w-full max-w-5xl">
            <Card variant="glass" className="hover-lift p-6">
              <CardContent className="p-0 space-y-3">
                <div className="text-3xl mb-2">🎯</div>
                <h3 className="text-lg font-semibold">Быстрый старт</h3>
                <p className="text-sm text-muted-foreground">
                  Присоединяйтесь к игре за секунды
                </p>
              </CardContent>
            </Card>
            
            <Card variant="glass" className="hover-lift p-6">
              <CardContent className="p-0 space-y-3">
                <div className="text-3xl mb-2">🧠</div>
                <h3 className="text-lg font-semibold">Стратегия</h3>
                <p className="text-sm text-muted-foreground">
                  Развивайте логику и интуицию
                </p>
              </CardContent>
            </Card>
            
            <Card variant="glass" className="hover-lift p-6">
              <CardContent className="p-0 space-y-3">
                <div className="text-3xl mb-2">👥</div>
                <h3 className="text-lg font-semibold">Сообщество</h3>
                <p className="text-sm text-muted-foreground">
                  Играйте с друзьями и новыми людьми
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </>
  )
}
