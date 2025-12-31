'use client'

import Link from 'next/link'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'next/navigation'
import { logout } from '@/lib/auth'
import { Button } from '@/components/ui/Button'

export function Header() {
  const { user } = useAuthStore()
  const router = useRouter()

  const handleLogout = () => {
    logout()
    router.push('/')
  }

  return (
    <header className="sticky top-0 z-50 w-full glass border-b border-border/50">
      <div className="container flex h-16 items-center justify-between px-4">
        <Link href="/" className="flex items-center space-x-2 group">
          <span className="text-2xl font-bold gradient-text group-hover:scale-110 transition-transform duration-200">
            🕵️ Mayoku
          </span>
        </Link>
        
        <nav className="flex items-center gap-2">
          <Link href="/lobby">
            <Button variant="ghost" size="sm" className="hover:bg-accent/50">
              Лобби
            </Button>
          </Link>
          <Link href="/profile">
            <Button variant="ghost" size="sm" className="hover:bg-accent/50">
              Профиль
            </Button>
          </Link>
          {user?.is_admin && (
            <Link href="/admin">
              <Button variant="ghost" size="sm" className="hover:bg-accent/50">
                Админ
              </Button>
            </Link>
          )}
          {user && (
            <Button
              onClick={handleLogout}
              variant="ghost"
              size="sm"
              className="text-muted-foreground hover:text-foreground"
            >
              Выход
            </Button>
          )}
        </nav>
      </div>
    </header>
  )
}

