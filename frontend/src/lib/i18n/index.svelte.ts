import { en } from './en'
import { ja } from './ja'

export type Locale = 'en' | 'ja'
export type LocalePref = '' | Locale
type Key = keyof typeof en
type Dict = { [K in Key]: string }

const dicts: Record<Locale, Dict> = { en, ja }

class I18n {
  locale = $state<Locale>('en')

  set(l: Locale) {
    this.locale = l
  }

  t(key: Key, params?: Record<string, string | number>): string {
    const v = dicts[this.locale][key] ?? en[key] ?? key
    if (!params) return v
    return v.replace(/\{(\w+)\}/g, (_, k) => {
      const p = params[k]
      return p === undefined ? `{${k}}` : String(p)
    })
  }
}

export const i18n = new I18n()

export function t(key: Key, params?: Record<string, string | number>): string {
  return i18n.t(key, params)
}
