'use client'

import { useState, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

interface FileUploadProps {
  accept?: string
  maxSize?: number // in MB
  onUpload: (file: File) => Promise<string>
  currentUrl?: string
  label: string
}

export function FileUpload({ accept, maxSize = 10, onUpload, currentUrl, label }: FileUploadProps) {
  const [uploading, setUploading] = useState(false)
  const [preview, setPreview] = useState<string | null>(currentUrl || null)
  const [error, setError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file size
    if (file.size > maxSize * 1024 * 1024) {
      setError(`文件大小不能超过 ${maxSize}MB`)
      return
    }

    setError(null)
    setUploading(true)

    try {
      // Create preview for images
      if (file.type.startsWith('image/')) {
        const reader = new FileReader()
        reader.onloadend = () => {
          setPreview(reader.result as string)
        }
        reader.readAsDataURL(file)
      }

      // Upload file
      const url = await onUpload(file)
      setPreview(url)
    } catch (err) {
      setError(err instanceof Error ? err.message : '上传失败')
      setPreview(null)
    } finally {
      setUploading(false)
    }
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="space-y-4">
          <label className="block text-sm font-medium">{label}</label>

          {preview && (
            <div className="relative w-full h-48 bg-gray-100 rounded-lg overflow-hidden">
              {preview.match(/\.(jpg|jpeg|png|gif|webp)$/i) ? (
                <img
                  src={preview}
                  alt="Preview"
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="flex items-center justify-center h-full">
                  <p className="text-sm text-gray-500">文件已上传</p>
                </div>
              )}
            </div>
          )}

          <div className="flex gap-2">
            <input
              ref={fileInputRef}
              type="file"
              accept={accept}
              onChange={handleFileChange}
              className="hidden"
            />
            <Button
              type="button"
              variant="outline"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
            >
              {uploading ? '上传中...' : preview ? '更换文件' : '选择文件'}
            </Button>
            {preview && (
              <Button
                type="button"
                variant="outline"
                onClick={() => setPreview(null)}
              >
                清除
              </Button>
            )}
          </div>

          {error && (
            <p className="text-sm text-red-500">{error}</p>
          )}

          <p className="text-xs text-gray-500">
            {accept && `支持格式: ${accept}`}
            {maxSize && ` | 最大 ${maxSize}MB`}
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
