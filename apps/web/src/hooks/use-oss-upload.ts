import { useState, useRef } from 'react'
import { apiClient } from '@/lib/api-client'

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
const MAX_SIZE = 10 * 1024 * 1024 // 10MB

function generateKey(dir: string, file: File): string {
  const ext = file.name.slice(file.name.lastIndexOf('.'))
  const uuid = crypto.randomUUID()
  const date = new Date().toISOString().slice(0, 10)
  return `${dir}${date}/${uuid}${ext}`
}

export function useOSSUpload() {
  const [progress, setProgress] = useState(0)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abortRef = useRef<XMLHttpRequest | null>(null)

  const upload = (file: File, purpose?: string): Promise<string> => {
    return new Promise(async (resolve, reject) => {
      setError(null)
      setProgress(0)

      if (!ALLOWED_TYPES.includes(file.type)) {
        const msg = '不支持的文件类型，请上传 JPG/PNG/GIF/WebP'
        setError(msg)
        reject(new Error(msg))
        return
      }
      if (file.size > MAX_SIZE) {
        const msg = '文件大小不能超过 10MB'
        setError(msg)
        reject(new Error(msg))
        return
      }

      const uploadViaLocalAPI = async () => {
        const result = await apiClient.uploadFile('/upload/image', file)
        return result.url
      }

      let policy
      try {
        policy = await apiClient.getOSSPolicy(purpose)
      } catch {
        try {
          setUploading(true)
          const localURL = await uploadViaLocalAPI()
          setUploading(false)
          setProgress(100)
          resolve(localURL)
          return
        } catch (localErr: any) {
          setUploading(false)
          const msg = localErr?.message ?? '获取上传凭证失败'
          setError(msg)
          reject(new Error(msg))
          return
        }
      }

      if (!policy?.host || !policy?.dir) {
        try {
          setUploading(true)
          const localURL = await uploadViaLocalAPI()
          setUploading(false)
          setProgress(100)
          resolve(localURL)
          return
        } catch (localErr: any) {
          setUploading(false)
          const msg = localErr?.message ?? '上传凭证无效'
          setError(msg)
          reject(new Error(msg))
          return
        }
      }

      const key = generateKey(policy.dir, file)
      const formData = new FormData()
      formData.append('key', key)
      formData.append('OSSAccessKeyId', policy.OSSAccessKeyId)
      formData.append('policy', policy.policy)
      formData.append('signature', policy.signature)
      formData.append('success_action_status', '200')
      formData.append('Content-Type', file.type)
      formData.append('file', file)

      setUploading(true)
      const xhr = new XMLHttpRequest()
      abortRef.current = xhr

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
          setProgress(Math.round((e.loaded / e.total) * 100))
        }
      }

      const fallbackToLocal = async () => {
        try {
          const localURL = await uploadViaLocalAPI()
          setUploading(false)
          abortRef.current = null
          setProgress(100)
          resolve(localURL)
        } catch (localErr: any) {
          setUploading(false)
          abortRef.current = null
          const msg = localErr?.message ?? '上传失败'
          setError(msg)
          reject(new Error(msg))
        }
      }

      xhr.onload = () => {
        setUploading(false)
        abortRef.current = null
        if (xhr.status === 200) {
          setProgress(100)
          resolve(`${policy.host}/${key}`)
        } else {
          void fallbackToLocal()
        }
      }

      xhr.onerror = () => {
        void fallbackToLocal()
      }

      xhr.timeout = 30000
      xhr.ontimeout = () => {
        void fallbackToLocal()
      }

      xhr.open('POST', policy.host)
      xhr.send(formData)
    })
  }

  const abort = () => {
    if (abortRef.current) {
      abortRef.current.abort()
      abortRef.current = null
      setUploading(false)
    }
  }

  return { upload, abort, progress, uploading, error }
}
