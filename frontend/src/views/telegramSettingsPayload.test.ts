import { describe, expect, it } from 'vitest'
import { pickTelegramSettings } from './telegramSettingsPayload'

describe('telegram settings payload', () => {
  it('keeps HasSecret markers only for secret Telegram settings', () => {
    const payload = pickTelegramSettings({
      telegramBotToken: '',
      telegramBotTokenHasSecret: 'true',
      telegramBackupPassphrase: '',
      telegramBackupPassphraseHasSecret: 'true',
      telegramBackupCron: '*/15 * * * *',
      telegramBackupCronHasSecret: 'true',
    })

    expect(payload.telegramBotTokenHasSecret).toBe('true')
    expect(payload.telegramBackupPassphraseHasSecret).toBe('true')
    expect(payload).not.toHaveProperty('telegramBackupCronHasSecret')
  })
})
