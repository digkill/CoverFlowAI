import { useRef, useState, useEffect, useCallback } from 'react'
import { Stage, Layer, Image as KonvaImage, Text, Transformer } from 'react-konva'
import useImage from 'use-image'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Textarea } from './ui/textarea'
import { Upload, Trash2, Sparkles, Type } from 'lucide-react'

interface CollageItem {
  id: string
  type: 'image'
  x: number
  y: number
  width: number
  height: number
  image: HTMLImageElement | null
  imageUrl: string
}

interface TextItem {
  id: string
  type: 'text'
  x: number
  y: number
  text: string
  fontSize: number
  fontFamily: string
  fill: string
  width: number
  height: number
}

type CanvasItem = CollageItem | TextItem

interface CollageEditorProps {
  onGenerate: (imageData: string, customPrompt?: string) => void
  isGenerating: boolean
}

const CANVAS_WIDTH = 1280
const CANVAS_HEIGHT = 720 // YouTube thumbnail standard size

export function CollageEditor({ onGenerate, isGenerating }: CollageEditorProps) {
  const stageRef = useRef<any>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const dropZoneRef = useRef<HTMLDivElement>(null)
  const [items, setItems] = useState<CanvasItem[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [customPrompt, setCustomPrompt] = useState<string>('')
  const [isDragging, setIsDragging] = useState(false)

  const handleFileUpload = useCallback((files: FileList | null) => {
    if (!files || files.length === 0) return

    Array.from(files).forEach((file) => {
      if (!file.type.startsWith('image/')) {
        alert('Пожалуйста, загрузите только изображения')
        return
      }

      const reader = new FileReader()
      reader.onload = (event) => {
        const imageUrl = event.target?.result as string
        const img = new window.Image()
        img.onload = () => {
          // Calculate size to fit within canvas
          const maxWidth = CANVAS_WIDTH / 2
          const maxHeight = CANVAS_HEIGHT / 2
          let width = img.width
          let height = img.height

          if (width > maxWidth || height > maxHeight) {
            const ratio = Math.min(maxWidth / width, maxHeight / height)
            width = width * ratio
            height = height * ratio
          }

          const newItem: CollageItem = {
            id: `item-${Date.now()}-${Math.random()}`,
            type: 'image',
            x: Math.random() * (CANVAS_WIDTH - width),
            y: Math.random() * (CANVAS_HEIGHT - height),
            width,
            height,
            image: img,
            imageUrl,
          }

          setItems((prev) => [...prev, newItem])
        }
        img.src = imageUrl
      }
      reader.readAsDataURL(file)
    })
  }, [])

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    handleFileUpload(e.target.files)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  // Drag and Drop handlers
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)

    const files = e.dataTransfer.files
    if (files.length > 0) {
      handleFileUpload(files)
    }
  }, [handleFileUpload])

  const handleAddText = () => {
    const newTextItem: TextItem = {
      id: `text-${Date.now()}-${Math.random()}`,
      type: 'text',
      x: CANVAS_WIDTH / 2 - 100,
      y: CANVAS_HEIGHT / 2 - 20,
      text: 'Новый текст',
      fontSize: 48,
      fontFamily: 'Arial',
      fill: '#000000',
      width: 200,
      height: 60,
    }
    setItems((prev) => [...prev, newTextItem])
    setSelectedId(newTextItem.id)
  }

  const handleDeleteSelected = () => {
    if (selectedId) {
      setItems((prev) => prev.filter((item) => item.id !== selectedId))
      setSelectedId(null)
    }
  }

  const handleGenerateCover = () => {
    const stage = stageRef.current
    if (!stage) return

    // Convert stage to data URL
    const dataURL = stage.toDataURL({
      pixelRatio: 2,
      quality: 1,
    })

    onGenerate(dataURL, customPrompt || undefined)
  }

  const handleStageClick = (e: any) => {
    const clickedOnEmpty = e.target === e.target.getStage()
    if (clickedOnEmpty) {
      setSelectedId(null)
    }
  }

  const selectedItem = items.find((item) => item.id === selectedId)

  return (
    <div className="space-y-4">
      {/* Controls */}
      <div className="flex gap-2">
        <Button
          variant="outline"
          onClick={() => fileInputRef.current?.click()}
          className="flex items-center gap-2"
        >
          <Upload className="w-4 h-4" />
          Загрузить изображения
        </Button>
        <Button
          variant="outline"
          onClick={handleAddText}
          className="flex items-center gap-2"
        >
          <Type className="w-4 h-4" />
          Добавить текст
        </Button>
      </div>
      <div className="flex gap-2">
        <Button
          variant="destructive"
          onClick={handleDeleteSelected}
          disabled={!selectedId}
          className="flex items-center gap-2"
        >
          <Trash2 className="w-4 h-4" />
          Удалить
        </Button>
        <Button
          onClick={handleGenerateCover}
          disabled={items.length === 0 || isGenerating}
          className="flex items-center gap-2"
        >
          <Sparkles className="w-4 h-4" />
          {isGenerating ? 'Генерация...' : 'Сгенерировать'}
        </Button>
      </div>

      {/* Custom Prompt Field */}
      <div className="space-y-2">
        <label htmlFor="custom-prompt" className="text-sm font-medium">
          Промпт для генерации (опционально):
        </label>
        <Textarea
          id="custom-prompt"
          placeholder="Опишите желаемую обложку или оставьте пустым для использования стандартного промпта"
          value={customPrompt}
          onChange={(e) => setCustomPrompt(e.target.value)}
          rows={2}
          className="resize-none"
        />
      </div>

         {/* Text Editor Panel */}
         {selectedItem?.type === 'text' && (
        <div className="border rounded-lg p-4 space-y-3 bg-muted/30">
          <h3 className="font-semibold">Редактирование текста</h3>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-sm text-muted-foreground">Текст:</label>
              <Input
                value={selectedItem.text}
                onChange={(e) => {
                  setItems((prev) =>
                    prev.map((i) =>
                      i.id === selectedItem.id
                        ? { ...i, text: e.target.value }
                        : i
                    )
                  )
                }}
              />
            </div>
            <div>
              <label className="text-sm text-muted-foreground">Размер шрифта:</label>
              <Input
                type="number"
                value={selectedItem.fontSize}
                onChange={(e) => {
                  setItems((prev) =>
                    prev.map((i) =>
                      i.id === selectedItem.id
                        ? { ...i, fontSize: parseInt(e.target.value) || 48 }
                        : i
                    )
                  )
                }}
                min="12"
                max="200"
              />
            </div>
            <div>
              <label className="text-sm text-muted-foreground">Цвет:</label>
              <Input
                type="color"
                value={selectedItem.fill}
                onChange={(e) => {
                  setItems((prev) =>
                    prev.map((i) =>
                      i.id === selectedItem.id
                        ? { ...i, fill: e.target.value }
                        : i
                    )
                  )
                }}
              />
            </div>
            <div>
              <label className="text-sm text-muted-foreground">Шрифт:</label>
              <select
                value={selectedItem.fontFamily}
                onChange={(e) => {
                  setItems((prev) =>
                    prev.map((i) =>
                      i.id === selectedItem.id
                        ? { ...i, fontFamily: e.target.value }
                        : i
                    )
                  )
                }}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="Arial">Arial</option>
                <option value="Helvetica">Helvetica</option>
                <option value="Times New Roman">Times New Roman</option>
                <option value="Courier New">Courier New</option>
                <option value="Verdana">Verdana</option>
                <option value="Georgia">Georgia</option>
                <option value="Impact">Impact</option>
              </select>
            </div>
          </div>
        </div>
      )}


      {/* Canvas with Drag and Drop Zone */}
      <div
        ref={dropZoneRef}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={`border rounded-lg overflow-hidden bg-muted/50 relative ${isDragging ? 'ring-2 ring-primary ring-offset-2' : ''
          }`}
      >
        {isDragging && (
          <div className="absolute inset-0 bg-primary/10 z-10 flex items-center justify-center pointer-events-none">
            <div className="bg-background border-2 border-dashed border-primary rounded-lg p-8">
              <Upload className="w-12 h-12 mx-auto mb-4 text-primary" />
              <p className="text-lg font-semibold">Перетащите изображения сюда</p>
            </div>
          </div>
        )}
        <Stage
          width={CANVAS_WIDTH}
          height={CANVAS_HEIGHT}
          ref={stageRef}
          onClick={handleStageClick}
          onTap={handleStageClick}
          style={{ maxWidth: '100%', height: 'auto' }}
        >
          <Layer>
            {items.map((item) => {
              if (item.type === 'image') {
                return (
                  <ImageItemComponent
                    key={item.id}
                    item={item}
                    isSelected={selectedId === item.id}
                    onSelect={() => setSelectedId(item.id)}
                    onUpdate={(updates) => {
                      setItems((prev) =>
                        prev.map((canvasItem) => {
                          if (canvasItem.id !== item.id) {
                            return canvasItem
                          }
                          if (canvasItem.type !== 'image') {
                            return canvasItem
                          }
                          return { ...canvasItem, ...updates }
                        })
                      )
                    }}
                  />
                )
              } else {
                return (
                  <TextItemComponent
                    key={item.id}
                    item={item}
                    isSelected={selectedId === item.id}
                    onSelect={() => setSelectedId(item.id)}
                    onUpdate={(updates) => {
                      setItems((prev) =>
                        prev.map((canvasItem) => {
                          if (canvasItem.id !== item.id) {
                            return canvasItem
                          }
                          if (canvasItem.type !== 'text') {
                            return canvasItem
                          }
                          return { ...canvasItem, ...updates }
                        })
                      )
                    }}
                  />
                )
              }
            })}
          </Layer>
        </Stage>
      </div>

   
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        multiple
        onChange={handleFileInputChange}
        className="hidden"
      />

      <p className="text-sm text-muted-foreground text-center">
        Размер холста: {CANVAS_WIDTH} × {CANVAS_HEIGHT}px (стандарт YouTube).
        Перетащите изображения на холст или добавьте текст.
      </p>
    </div>
  )
}

// Image Component
interface ImageItemComponentProps {
  item: CollageItem
  isSelected: boolean
  onSelect: () => void
  onUpdate: (updates: Partial<CollageItem>) => void
}

function ImageItemComponent({
  item,
  isSelected,
  onSelect,
  onUpdate,
}: ImageItemComponentProps) {
  const [image] = useImage(item.imageUrl)
  const shapeRef = useRef<any>(null)
  const trRef = useRef<any>(null)

  useEffect(() => {
    if (isSelected && trRef.current && shapeRef.current) {
      trRef.current.nodes([shapeRef.current])
      trRef.current.getLayer()?.batchDraw()
    }
  }, [isSelected])

  return (
    <>
      <KonvaImage
        ref={shapeRef}
        image={image}
        x={item.x}
        y={item.y}
        width={item.width}
        height={item.height}
        draggable
        onClick={onSelect}
        onTap={onSelect}
        onDragEnd={(e) => {
          onUpdate({
            x: e.target.x(),
            y: e.target.y(),
          })
        }}
        onTransformEnd={(e) => {
          const node = e.target
          const scaleX = node.scaleX()
          const scaleY = node.scaleY()

          onUpdate({
            x: node.x(),
            y: node.y(),
            width: Math.max(50, node.width() * scaleX),
            height: Math.max(50, node.height() * scaleY),
          })

          node.scaleX(1)
          node.scaleY(1)
        }}
        shadowBlur={isSelected ? 10 : 0}
        shadowColor="rgba(0,0,0,0.3)"
      />
      {isSelected && (
        <Transformer
          ref={trRef}
          boundBoxFunc={(oldBox, newBox) => {
            if (Math.abs(newBox.width) < 50 || Math.abs(newBox.height) < 50) {
              return oldBox
            }
            return newBox
          }}
        />
      )}
    </>
  )
}

// Text Component
interface TextItemComponentProps {
  item: TextItem
  isSelected: boolean
  onSelect: () => void
  onUpdate: (updates: Partial<TextItem>) => void
}

function TextItemComponent({
  item,
  isSelected,
  onSelect,
  onUpdate,
}: TextItemComponentProps) {
  const textRef = useRef<any>(null)
  const trRef = useRef<any>(null)

  useEffect(() => {
    if (isSelected && trRef.current && textRef.current) {
      trRef.current.nodes([textRef.current])
      trRef.current.getLayer()?.batchDraw()
    }
  }, [isSelected])

  return (
    <>
      <Text
        ref={textRef}
        x={item.x}
        y={item.y}
        text={item.text}
        fontSize={item.fontSize}
        fontFamily={item.fontFamily}
        fill={item.fill}
        draggable
        onClick={onSelect}
        onTap={onSelect}
        onDragEnd={(e) => {
          onUpdate({
            x: e.target.x(),
            y: e.target.y(),
          })
        }}
        onTransformEnd={(e) => {
          const node = e.target
          const scaleX = node.scaleX()
          onUpdate({
            x: node.x(),
            y: node.y(),
            fontSize: Math.max(12, item.fontSize * scaleX),
          })

          node.scaleX(1)
          node.scaleY(1)
        }}
        shadowBlur={isSelected ? 5 : 0}
        shadowColor="rgba(0,0,0,0.3)"
      />
      {isSelected && (
        <Transformer
          ref={trRef}
          boundBoxFunc={(oldBox, newBox) => {
            if (Math.abs(newBox.width) < 20 || Math.abs(newBox.height) < 20) {
              return oldBox
            }
            return newBox
          }}
        />
      )}
    </>
  )
}
