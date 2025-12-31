'use client'

import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { Header } from '@/components/layout/Header'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { api } from '@/lib/api'
import { uploadFile } from '@/lib/upload'
import type { Deck, Location } from '@/types'
import { useAuthStore } from '@/stores/auth'

interface LocationForm {
  name: string
  image_url: string
  roles: string[]
}

export default function DeckBuilderPage() {
  const router = useRouter()
  const { user } = useAuthStore()
  const [deckName, setDeckName] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [locations, setLocations] = useState<LocationForm[]>([
    { name: '', image_url: '', roles: [] }
  ])
  const [uploadingImage, setUploadingImage] = useState<number | null>(null)
  const [newRole, setNewRole] = useState<Record<number, string>>({})

  const createDeckMutation = useMutation({
    mutationFn: async (data: {
      name: string
      is_public: boolean
      locations: Array<{
        name: string
        image_url: string
        roles: string[]
      }>
    }) => {
      return api.post<Deck>('/api/decks', data)
    },
    onSuccess: () => {
      router.push('/profile')
    },
  })

  const handleAddLocation = () => {
    setLocations([...locations, { name: '', image_url: '', roles: [] }])
  }

  const handleRemoveLocation = (index: number) => {
    setLocations(locations.filter((_, i) => i !== index))
  }

  const handleLocationChange = (index: number, field: keyof LocationForm, value: string) => {
    const newLocations = [...locations]
    newLocations[index] = { ...newLocations[index], [field]: value }
    setLocations(newLocations)
  }

  const handleImageUpload = async (index: number, file: File) => {
    if (!file.type.startsWith('image/')) {
      alert('Пожалуйста, выберите изображение')
      return
    }

    setUploadingImage(index)
    try {
      const url = await uploadFile(file)
      handleLocationChange(index, 'image_url', url)
    } catch (error) {
      console.error('Upload failed:', error)
      alert('Ошибка загрузки изображения')
    } finally {
      setUploadingImage(null)
    }
  }

  const handleAddRole = (locationIndex: number) => {
    const role = newRole[locationIndex]?.trim()
    if (!role) return

    const newLocations = [...locations]
    if (!newLocations[locationIndex].roles) {
      newLocations[locationIndex].roles = []
    }
    newLocations[locationIndex].roles = [...newLocations[locationIndex].roles, role]
    setLocations(newLocations)
    setNewRole({ ...newRole, [locationIndex]: '' })
  }

  const handleRemoveRole = (locationIndex: number, roleIndex: number) => {
    const newLocations = [...locations]
    newLocations[locationIndex].roles = newLocations[locationIndex].roles.filter(
      (_, i) => i !== roleIndex
    )
    setLocations(newLocations)
  }

  const handleSubmit = () => {
    // Валидация
    if (!deckName.trim()) {
      alert('Введите название колоды')
      return
    }

    if (locations.length === 0) {
      alert('Добавьте хотя бы одну локацию')
      return
    }

    for (const location of locations) {
      if (!location.name.trim()) {
        alert('Все локации должны иметь название')
        return
      }
      if (!location.image_url) {
        alert('Все локации должны иметь изображение')
        return
      }
      if (location.roles.length === 0) {
        alert('Все локации должны иметь хотя бы одну роль')
        return
      }
    }

    createDeckMutation.mutate({
      name: deckName,
      is_public: isPublic,
      locations: locations.map(loc => ({
        name: loc.name,
        image_url: loc.image_url,
        roles: loc.roles,
      })),
    })
  }

  return (
    <>
      <Header />
      <main className="container mx-auto px-4 py-12 max-w-4xl">
        <div className="space-y-6">
          <Card variant="glass">
            <CardHeader>
              <CardTitle className="text-3xl">Создать колоду</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Deck Name */}
              <div className="space-y-2">
                <label className="text-sm font-medium">Название колоды</label>
                <Input
                  value={deckName}
                  onChange={(e) => setDeckName(e.target.value)}
                  placeholder="Например: Популярные места"
                  className="w-full"
                />
              </div>

              {/* Public Toggle */}
              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  id="isPublic"
                  checked={isPublic}
                  onChange={(e) => setIsPublic(e.target.checked)}
                  className="w-4 h-4 rounded border-input"
                />
                <label htmlFor="isPublic" className="text-sm font-medium cursor-pointer">
                  Сделать колоду публичной
                </label>
              </div>

              {/* Locations */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold">Локации</h3>
                  <Button onClick={handleAddLocation} variant="secondary" size="sm" className="hover-lift">
                    + Добавить локацию
                  </Button>
                </div>

                {locations.map((location, index) => (
                  <Card key={index} variant="elevated" className="hover-lift">
                    <CardContent className="p-6 space-y-4">
                      <div className="flex items-center justify-between">
                        <h4 className="font-semibold">Локация {index + 1}</h4>
                        {locations.length > 1 && (
                          <Button
                            variant="danger"
                            size="sm"
                            onClick={() => handleRemoveLocation(index)}
                          >
                            Удалить
                          </Button>
                        )}
                      </div>

                      {/* Location Name */}
                      <div className="space-y-2">
                        <label className="text-sm font-medium">Название локации</label>
                        <Input
                          value={location.name}
                          onChange={(e) => handleLocationChange(index, 'name', e.target.value)}
                          placeholder="Например: Ресторан"
                          className="w-full"
                        />
                      </div>

                      {/* Image Upload */}
                      <div className="space-y-2">
                        <label className="text-sm font-medium">Изображение</label>
                        <div className="flex items-center gap-4">
                          <input
                            type="file"
                            accept="image/*"
                            onChange={(e) => {
                              const file = e.target.files?.[0]
                              if (file) handleImageUpload(index, file)
                            }}
                            className="hidden"
                            id={`image-${index}`}
                            disabled={uploadingImage === index}
                          />
                          <label
                            htmlFor={`image-${index}`}
                            className="cursor-pointer"
                          >
                            <Button
                              type="button"
                              variant="secondary"
                              size="sm"
                              disabled={uploadingImage === index}
                              className="hover-lift"
                            >
                              {uploadingImage === index ? 'Загрузка...' : 'Выбрать изображение'}
                            </Button>
                          </label>
                          {location.image_url && (
                            <div className="flex-1">
                              <img
                                src={location.image_url}
                                alt={location.name || 'Preview'}
                                className="w-32 h-32 object-cover rounded-lg border border-border"
                                onError={(e) => {
                                  // Fallback если изображение не загрузилось
                                  console.error('Failed to load image:', location.image_url)
                                }}
                              />
                            </div>
                          )}
                        </div>
                      </div>

                      {/* Roles */}
                      <div className="space-y-2">
                        <label className="text-sm font-medium">Роли</label>
                        <div className="flex gap-2">
                          <Input
                            value={newRole[index] || ''}
                            onChange={(e) => setNewRole({ ...newRole, [index]: e.target.value })}
                            placeholder="Новая роль"
                            className="flex-1"
                            onKeyPress={(e) => {
                              if (e.key === 'Enter') {
                                e.preventDefault()
                                handleAddRole(index)
                              }
                            }}
                          />
                          <Button
                            onClick={() => handleAddRole(index)}
                            variant="secondary"
                            size="sm"
                            className="hover-lift"
                          >
                            Добавить
                          </Button>
                        </div>
                        <div className="flex flex-wrap gap-2 mt-2">
                          {location.roles.map((role, roleIndex) => (
                            <div
                              key={roleIndex}
                              className="flex items-center gap-2 px-3 py-1 rounded-full bg-primary/20 text-primary text-sm"
                            >
                              <span>{role}</span>
                              <button
                                onClick={() => handleRemoveRole(index, roleIndex)}
                                className="hover:text-destructive"
                              >
                                ×
                              </button>
                            </div>
                          ))}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {/* Submit */}
              <div className="flex gap-4 pt-4">
                <Button
                  onClick={handleSubmit}
                  disabled={createDeckMutation.isPending}
                  size="lg"
                  className="flex-1 hover-lift glow"
                >
                  {createDeckMutation.isPending ? 'Создание...' : 'Создать колоду'}
                </Button>
                <Button
                  onClick={() => router.back()}
                  variant="secondary"
                  size="lg"
                  className="hover-lift"
                >
                  Отмена
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </main>
    </>
  )
}

