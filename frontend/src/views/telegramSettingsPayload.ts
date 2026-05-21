export type TelegramSettingsMap = Record<string, string>

export const telegramSettingKeys = [
  'telegramEnabled',
  'telegramBotToken',
  'telegramChatID',
  'telegramProxyURL',
  'telegramProxyUsername',
  'telegramProxyPassword',
  'telegramCpuThreshold',
  'telegramNotifyCpu',
  'telegramReport',
  'telegramReportCron',
  'telegramBackupEnabled',
  'telegramBackupPassphrase',
  'telegramBackupCron',
  'telegramBackupExcludeTables',
  'telegramBackupMaxSizeMB',
]

const telegramSecretSettingKeys = [
  'telegramBotToken',
  'telegramProxyURL',
  'telegramProxyUsername',
  'telegramProxyPassword',
  'telegramBackupPassphrase',
]

export const pickTelegramSettings = (source: TelegramSettingsMap): TelegramSettingsMap => {
  const picked: TelegramSettingsMap = {}
  for (const key of telegramSettingKeys) {
    picked[key] = String(source[key] ?? '')
  }
  for (const key of telegramSecretSettingKeys) {
    picked[key + 'HasSecret'] = String(source[key + 'HasSecret'] ?? 'false')
  }
  return picked
}
