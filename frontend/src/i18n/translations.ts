export const translations = {
  ru: {
    // Common
    common: {
      loading: 'Загрузка...',
      error: 'Ошибка',
      success: 'Успешно',
      cancel: 'Отмена',
      save: 'Сохранить',
      delete: 'Удалить',
      edit: 'Редактировать',
      close: 'Закрыть',
    },
    // Navigation
    nav: {
      home: 'Главная',
      generate: 'Генерация',
      myCovers: 'Мои обложки',
      settings: 'Настройки',
    },
    // Auth
    auth: {
      signIn: 'Войти',
      signOut: 'Выйти',
      signInWithGoogle: 'Войти через Google',
      welcome: 'Добро пожаловать',
      notSignedIn: 'Вы не авторизованы',
    },
    // Editor
    editor: {
      title: 'Редактор коллажа',
      uploadImages: 'Загрузить изображения',
      addText: 'Добавить текст',
      delete: 'Удалить',
      generate: 'Сгенерировать обложку',
      generating: 'Генерация...',
      customPrompt: 'Промпт для генерации (опционально)',
      promptPlaceholder: 'Опишите желаемую обложку или оставьте пустым для использования стандартного промпта',
      canvasSize: 'Размер холста: {width} × {height}px (стандарт YouTube). Перетащите изображения на холст или добавьте текст.',
      dragDrop: 'Перетащите изображения сюда',
      // Text editor
      textEditor: {
        title: 'Редактирование текста',
        text: 'Текст:',
        fontSize: 'Размер шрифта:',
        color: 'Цвет:',
        font: 'Шрифт:',
      },
    },
    // Results
    results: {
      title: 'Сгенерированная обложка',
      download: 'Скачать обложку',
      clear: 'Очистить',
      loading: 'Генерация обложки с помощью {provider}...',
      placeholder: 'Создайте коллаж и нажмите "Сгенерировать обложку"',
      error: 'Ошибка загрузки изображения',
    },
    // Limits & Payments
    limits: {
      title: 'Лимиты генераций',
      freeGenerations: 'Бесплатные генерации',
      remaining: 'Осталось генераций: {count}',
      noRemaining: 'У вас не осталось бесплатных генераций',
      buyMore: 'Купить генерации',
      pricing: {
        title: 'Тарифы',
        free: 'Бесплатно',
        perDay: 'в день',
        pack1: 'Пакет 1: 10 генераций',
        pack2: 'Пакет 2: 30 генераций',
        pack3: 'Пакет 3: 100 генераций',
        popular: 'Популярный',
      },
    },
    // Errors
    errors: {
      authRequired: 'Для генерации необходимо войти в систему',
      noGenerationsLeft: 'У вас не осталось генераций. Купите дополнительный пакет.',
      paymentFailed: 'Ошибка при оплате',
      generationFailed: 'Ошибка при генерации обложки',
    },
  },
  en: {
    // Common
    common: {
      loading: 'Loading...',
      error: 'Error',
      success: 'Success',
      cancel: 'Cancel',
      save: 'Save',
      delete: 'Delete',
      edit: 'Edit',
      close: 'Close',
    },
    // Navigation
    nav: {
      home: 'Home',
      generate: 'Generate',
      myCovers: 'My Covers',
      settings: 'Settings',
    },
    // Auth
    auth: {
      signIn: 'Sign In',
      signOut: 'Sign Out',
      signInWithGoogle: 'Sign in with Google',
      welcome: 'Welcome',
      notSignedIn: 'You are not signed in',
    },
    // Editor
    editor: {
      title: 'Collage Editor',
      uploadImages: 'Upload Images',
      addText: 'Add Text',
      delete: 'Delete',
      generate: 'Generate Cover',
      generating: 'Generating...',
      customPrompt: 'Generation Prompt (optional)',
      promptPlaceholder: 'Describe the desired cover or leave empty to use the default prompt',
      canvasSize: 'Canvas size: {width} × {height}px (YouTube standard). Drag images onto the canvas or add text.',
      dragDrop: 'Drop images here',
      // Text editor
      textEditor: {
        title: 'Text Editing',
        text: 'Text:',
        fontSize: 'Font Size:',
        color: 'Color:',
        font: 'Font:',
      },
    },
    // Results
    results: {
      title: 'Generated Cover',
      download: 'Download Cover',
      clear: 'Clear',
      loading: 'Generating cover with {provider}...',
      placeholder: 'Create a collage and click "Generate Cover"',
      error: 'Failed to load image',
    },
    // Limits & Payments
    limits: {
      title: 'Generation Limits',
      freeGenerations: 'Free Generations',
      remaining: 'Generations remaining: {count}',
      noRemaining: 'You have no free generations left',
      buyMore: 'Buy Generations',
      pricing: {
        title: 'Pricing',
        free: 'Free',
        perDay: 'per day',
        pack1: 'Pack 1: 10 generations',
        pack2: 'Pack 2: 30 generations',
        pack3: 'Pack 3: 100 generations',
        popular: 'Popular',
      },
    },
    // Errors
    errors: {
      authRequired: 'You must sign in to generate covers',
      noGenerationsLeft: 'You have no generations left. Purchase an additional pack.',
      paymentFailed: 'Payment failed',
      generationFailed: 'Failed to generate cover',
    },
  },
  packages: {
    title: { ru: 'Выберите пакет генераций', en: 'Choose Generation Package' },
    currency: { ru: 'Валюта', en: 'Currency' },
    generations: { ru: 'генераций', en: 'generations' },
    perGeneration: { ru: 'генерация', en: 'per generation' },
    popular: { ru: 'Популярный', en: 'Popular' },
    highQuality: { ru: 'Высокое качество', en: 'High Quality' },
    noWatermark: { ru: 'Без водяных знаков', en: 'No Watermarks' },
    buy: { ru: 'Купить', en: 'Buy' },
    freePlan: { ru: 'Бесплатный план: 1 генерация в день', en: 'Free Plan: 1 generation per day' },
  },
  generations: {
    remaining: { ru: 'Генераций:', en: 'Generations:' },
  },
}

export type Language = 'ru' | 'en'
export type TranslationKey = string

export function t(key: TranslationKey, lang: Language = 'ru', params?: Record<string, string | number>): string {
  const keys = key.split('.')
  let value: any = translations[lang]
  
  for (const k of keys) {
    value = value?.[k]
  }
  
  if (typeof value !== 'string') {
    return key
  }
  
  // Replace parameters
  if (params) {
    return value.replace(/\{(\w+)\}/g, (match, param) => {
      return String(params[param] ?? match)
    })
  }
  
  return value
}

