import { useState, useEffect } from 'react'
import { Card } from './ui/card'
import { Button } from './ui/button'
import { X, Check, Sparkles } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'

interface Package {
  type: string
  name: string
  count: number
  price_usd: number
  price_rub: number
  popular: boolean
}

interface PackagesModalProps {
  isOpen: boolean
  onClose: () => void
  onPurchase: (packageType: string, currency: string) => void
}

export function PackagesModal({ isOpen, onClose, onPurchase }: PackagesModalProps) {
  const [packages, setPackages] = useState<Package[]>([])
  const [selectedCurrency, setSelectedCurrency] = useState<'USD' | 'RUB'>('RUB')
  const { t } = useLanguage()

  useEffect(() => {
    if (isOpen) {
      fetch('http://localhost:8080/api/packages', {
        credentials: 'include'
      })
        .then(res => res.json())
        .then(data => setPackages(data.packages || []))
        .catch(err => console.error('Failed to load packages:', err))
    }
  }, [isOpen])

  if (!isOpen) return null

  const handlePurchase = (packageType: string) => {
    onPurchase(packageType, selectedCurrency)
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold">
              {t('packages.title', 'Выберите пакет генераций')}
            </h2>
            <button
              onClick={onClose}
              className="p-2 hover:bg-muted rounded-md transition-colors"
            >
              <X className="h-5 w-5" />
            </button>
          </div>

          {/* Currency selector */}
          <div className="flex items-center gap-4 mb-6">
            <span className="text-sm font-medium">
              {t('packages.currency', 'Валюта')}:
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => setSelectedCurrency('RUB')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  selectedCurrency === 'RUB'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted hover:bg-muted/80'
                }`}
              >
                ₽ RUB
              </button>
              <button
                onClick={() => setSelectedCurrency('USD')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  selectedCurrency === 'USD'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted hover:bg-muted/80'
                }`}
              >
                $ USD
              </button>
            </div>
          </div>

          {/* Packages grid */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {packages.map((pkg) => (
              <Card
                key={pkg.type}
                className={`p-6 relative ${
                  pkg.popular
                    ? 'ring-2 ring-primary bg-primary/5'
                    : ''
                }`}
              >
                {pkg.popular && (
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-primary text-primary-foreground text-xs font-semibold px-3 py-1 rounded-full">
                    {t('packages.popular', 'Популярный')}
                  </div>
                )}
                
                <div className="text-center mb-4">
                  <h3 className="text-xl font-bold mb-2">{pkg.name}</h3>
                  <div className="flex items-baseline justify-center gap-2">
                    <span className="text-3xl font-bold">
                      {selectedCurrency === 'RUB' 
                        ? `${pkg.price_rub}₽`
                        : `$${pkg.price_usd}`
                      }
                    </span>
                  </div>
                  <p className="text-sm text-muted-foreground mt-2">
                    {pkg.count} {t('packages.generations', 'генераций')}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {selectedCurrency === 'RUB'
                      ? `${(pkg.price_rub / pkg.count).toFixed(1)}₽`
                      : `$${(pkg.price_usd / pkg.count).toFixed(2)}`
                    } / {t('packages.perGeneration', 'генерация')}
                  </p>
                </div>

                <ul className="space-y-2 mb-6">
                  <li className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-primary flex-shrink-0" />
                    <span>{pkg.count} {t('packages.generations', 'генераций')}</span>
                  </li>
                  <li className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-primary flex-shrink-0" />
                    <span>{t('packages.highQuality', 'Высокое качество')}</span>
                  </li>
                  <li className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-primary flex-shrink-0" />
                    <span>{t('packages.noWatermark', 'Без водяных знаков')}</span>
                  </li>
                </ul>

                <Button
                  onClick={() => handlePurchase(pkg.type)}
                  className="w-full"
                  variant={pkg.popular ? 'default' : 'outline'}
                >
                  <Sparkles className="h-4 w-4 mr-2" />
                  {t('packages.buy', 'Купить')}
                </Button>
              </Card>
            ))}
          </div>

          <div className="mt-6 text-center text-sm text-muted-foreground">
            <p>
              {t('packages.freePlan', 'Бесплатный план: 1 генерация в день')}
            </p>
          </div>
        </div>
      </Card>
    </div>
  )
}

