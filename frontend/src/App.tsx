import { useState } from 'react'
import { CollageEditor } from './components/CollageEditor'
import { PackagesModal } from './components/PackagesModal'
import { Button } from './components/ui/button'
import { Card } from './components/ui/card'
import { Upload, Sparkles, LogIn, LogOut, User, ShoppingCart } from 'lucide-react'
import axios from 'axios'
import { useAuth } from './contexts/AuthContext'
import { useLanguage } from './contexts/LanguageContext'

interface GeneratedCover {
  url: string
  id: string
}

type Provider = 'nanobanana' | 'openai'

function App() {
  const [generatedCover, setGeneratedCover] = useState<GeneratedCover | null>(null)
  const [isGenerating, setIsGenerating] = useState(false)
  const [provider] = useState<Provider>('nanobanana')
  const [showPackages, setShowPackages] = useState(false)
  const { user, isLoading: authLoading, signIn, signOut, refreshUser } = useAuth()
  const { t, language, setLanguage } = useLanguage()

  const handleGenerate = async (collageImageData: string, customPrompt?: string) => {
    setIsGenerating(true)
    try {
      const response = await axios.post('http://localhost:8080/api/generate-cover', {
        image: collageImageData, // base64 encoded image
        provider: provider,
        prompt: customPrompt, // optional custom prompt
      }, {
        headers: {
          'Content-Type': 'application/json',
        },
      })

      setGeneratedCover({
        url: response.data.image_url,
        id: response.data.id,
      })
    } catch (error: any) {
      console.error('Error generating cover:', error)
      
      let errorMessage = 'Ошибка при генерации обложки'
      
      if (error.response?.data?.details) {
        const details = error.response.data.details
        
        if (details.includes('IMGBB_API_KEY not set')) {
          errorMessage = 'Ошибка: Не установлен IMGBB_API_KEY. Получите ключ на https://api.imgbb.com/ и добавьте в .env файл backend.'
        } else if (details.includes('API key not configured')) {
          errorMessage = 'Ошибка: API ключ не настроен. Проверьте .env файл backend.'
        } else if (details.includes('authentication failed')) {
          errorMessage = 'Ошибка: Неверный API ключ. Проверьте правильность ключа в .env файле.'
        } else if (details.includes('insufficient account balance')) {
          errorMessage = 'Ошибка: Недостаточно средств на аккаунте API.'
        } else if (details.includes('rate limit exceeded')) {
          errorMessage = 'Ошибка: Превышен лимит запросов. Попробуйте позже.'
        } else {
          errorMessage = `Ошибка: ${details}`
        }
      } else if (error.response?.status === 402) {
        // Payment required - no generations left
        errorMessage = error.response.data?.message || 'У вас закончились генерации. Пожалуйста, купите пакет для продолжения.'
        setShowPackages(true)
      } else if (error.response?.data?.error) {
        errorMessage = `Ошибка: ${error.response.data.error}`
      } else if (error.message) {
        errorMessage = `Ошибка: ${error.message}`
      }
      
      if (error.response?.status !== 402) {
        alert(errorMessage)
      }
    } finally {
      setIsGenerating(false)
    }
  }

  const handlePurchase = async (packageType: string, currency: string) => {
    try {
      const response = await fetch('http://localhost:8080/api/payment/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          package_type: packageType,
          currency: currency,
        }),
      })

      if (response.ok) {
        const data = await response.json()
        // Redirect to payment URL
        window.open(data.payment_url, '_blank')
        // Refresh user after a delay (in real app, use webhook)
        setTimeout(() => {
          refreshUser()
        }, 5000)
      } else {
        const error = await response.json()
        alert(`Ошибка создания платежа: ${error.error}`)
      }
    } catch (error) {
      console.error('Purchase failed:', error)
      alert('Ошибка при создании платежа')
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-background via-secondary/20 to-background">
      {/* Auth Header */}
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <h1 className="text-xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
                CoverFlow AI
              </h1>
            </div>
            
            <div className="flex items-center gap-4">
              {/* Language Switcher */}
              <div className="flex items-center gap-2 border rounded-md p-1">
                <button
                  onClick={() => setLanguage('ru')}
                  className={`px-2 py-1 text-sm rounded transition-colors ${
                    language === 'ru'
                      ? 'bg-primary text-primary-foreground'
                      : 'hover:bg-muted'
                  }`}
                >
                  RU
                </button>
                <button
                  onClick={() => setLanguage('en')}
                  className={`px-2 py-1 text-sm rounded transition-colors ${
                    language === 'en'
                      ? 'bg-primary text-primary-foreground'
                      : 'hover:bg-muted'
                  }`}
                >
                  EN
                </button>
              </div>

              {/* Auth Section */}
              {authLoading ? (
                <div className="text-sm text-muted-foreground">{t('common.loading')}</div>
              ) : user ? (
                <div className="flex items-center gap-3">
                  <div className="flex items-center gap-2">
                    {user.picture ? (
                      <img
                        src={user.picture}
                        alt={user.name}
                        className="w-8 h-8 rounded-full border"
                      />
                    ) : (
                      <div className="w-8 h-8 rounded-full border flex items-center justify-center bg-muted">
                        <User className="w-4 h-4" />
                      </div>
                    )}
                    <div className="hidden sm:block text-sm">
                      <div className="font-medium">{user.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {t('generations.remaining')} {user.generations_remaining || 0}
                      </div>
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowPackages(true)}
                    className="flex items-center gap-2"
                  >
                    <ShoppingCart className="w-4 h-4" />
                    <span className="hidden sm:inline">{t('packages.buy')}</span>
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={signOut}
                    className="flex items-center gap-2"
                  >
                    <LogOut className="w-4 h-4" />
                    <span className="hidden sm:inline">{t('auth.signOut')}</span>
                  </Button>
                </div>
              ) : (
                <Button
                  onClick={signIn}
                  className="flex items-center gap-2"
                >
                  <LogIn className="w-4 h-4" />
                  {t('auth.signInWithGoogle')}
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-2 bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent">
            CoverFlow AI
          </h1>
          <p className="text-muted-foreground">
            {t('editor.title')}
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card className="p-6">
            <div className="mb-4 flex items-center gap-2">
              <Upload className="w-5 h-5" />
              <h2 className="text-2xl font-semibold">{t('editor.title')}</h2>
            </div>
            <CollageEditor onGenerate={handleGenerate} isGenerating={isGenerating} />
          </Card>

          <Card className="p-6">
            <div className="mb-4 flex items-center gap-2">
              <Sparkles className="w-5 h-5" />
              <h2 className="text-2xl font-semibold">{t('results.title')}</h2>
            </div>
            <div className="space-y-4">
              {generatedCover ? (
                <div className="space-y-4">
                  <div className="aspect-video bg-muted rounded-lg overflow-hidden border relative">
                    <img
                      src={generatedCover.url}
                      alt="Generated cover"
                      className="w-full h-full object-contain"
                      onError={(e) => {
                        console.error('Failed to load image:', generatedCover.url)
                        const target = e.target as HTMLImageElement
                        target.style.display = 'none'
                        const errorDiv = document.createElement('div')
                        errorDiv.className = 'absolute inset-0 flex items-center justify-center text-muted-foreground'
                        errorDiv.textContent = t('results.error')
                        target.parentElement?.appendChild(errorDiv)
                      }}
                      crossOrigin="anonymous"
                    />
                  </div>
                  <div className="flex gap-2">
                    <Button
                      onClick={async () => {
                        try {
                          // Try to fetch image and create download
                          const response = await fetch(generatedCover.url)
                          const blob = await response.blob()
                          const url = window.URL.createObjectURL(blob)
                          const link = document.createElement('a')
                          link.href = url
                          link.download = `cover-${generatedCover.id}.png`
                          document.body.appendChild(link)
                          link.click()
                          document.body.removeChild(link)
                          window.URL.revokeObjectURL(url)
                        } catch (error) {
                          console.error('Download error:', error)
                          // Fallback: open in new tab
                          window.open(generatedCover.url, '_blank')
                        }
                      }}
                      className="flex-1"
                    >
                      {t('results.download')}
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => setGeneratedCover(null)}
                    >
                      {t('results.clear')}
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="aspect-video bg-muted rounded-lg border-2 border-dashed flex items-center justify-center">
                  <p className="text-muted-foreground text-center px-4">
                    {isGenerating
                      ? t('results.loading', { provider: provider === 'nanobanana' ? 'Nano Banana' : 'OpenAI' })
                      : t('results.placeholder')}
                  </p>
                </div>
              )}
            </div>
          </Card>
        </div>
      </div>

      <PackagesModal
        isOpen={showPackages}
        onClose={() => setShowPackages(false)}
        onPurchase={handlePurchase}
      />
    </div>
  )
}

export default App
