import type { Metadata } from 'next'
import './globals.css'
import { Providers } from './providers'
import { AuthGuard } from '@/components/auth/AuthGuard'

export const metadata: Metadata = {
  title: 'Mayoku - Spy Game',
  description: 'Современная игра "Шпион" для Telegram Mini Apps',
  viewport: 'width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ru">
      <body>
        <Providers>
          <AuthGuard>{children}</AuthGuard>
        </Providers>
      </body>
    </html>
  )
}
