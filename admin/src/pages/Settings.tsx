import { useState } from 'react'
import { toast } from 'react-toastify'
import { Settings as SettingsIcon, Save } from 'lucide-react'

export function Settings() {
  const [settings, setSettings] = useState({
    siteName: 'MyPila',
    maintenanceMode: false,
    allowRegistration: true,
    defaultCurrency: 'BRL',
    emailNotifications: true
  })

  const handleSave = () => {
    toast.success('Configurações salvas com sucesso!')
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Configurações do Sistema</h1>

      <div className="bg-white rounded-lg shadow-md p-6 max-w-2xl">
        <div className="flex items-center gap-3 mb-6">
          <SettingsIcon className="text-blue-600" size={24} />
          <h2 className="text-lg font-semibold text-gray-900">Gerais</h2>
        </div>

        <div className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Nome do Site
            </label>
            <input
              type="text"
              value={settings.siteName}
              onChange={(e) => setSettings({ ...settings, siteName: e.target.value })}
              className="input"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Moeda Padrão
            </label>
            <select
              value={settings.defaultCurrency}
              onChange={(e) => setSettings({ ...settings, defaultCurrency: e.target.value })}
              className="input"
            >
              <option value="BRL">Real Brasileiro (R$)</option>
              <option value="USD">Dólar Americano ($)</option>
              <option value="EUR">Euro (€)</option>
            </select>
          </div>

          <div className="flex items-center justify-between py-3 border-t border-gray-200">
            <div>
              <h3 className="text-sm font-medium text-gray-900">Modo de Manutenção</h3>
              <p className="text-sm text-gray-500">Bloquear acesso ao site para manutenção</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.maintenanceMode}
                onChange={(e) => setSettings({ ...settings, maintenanceMode: e.target.checked })}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>

          <div className="flex items-center justify-between py-3 border-t border-gray-200">
            <div>
              <h3 className="text-sm font-medium text-gray-900">Permitir Novos Registros</h3>
              <p className="text-sm text-gray-500">Usuários podem criar novas contas</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.allowRegistration}
                onChange={(e) => setSettings({ ...settings, allowRegistration: e.target.checked })}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>

          <div className="flex items-center justify-between py-3 border-t border-gray-200">
            <div>
              <h3 className="text-sm font-medium text-gray-900">Notificações por Email</h3>
              <p className="text-sm text-gray-500">Enviar notificações importantes por email</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.emailNotifications}
                onChange={(e) => setSettings({ ...settings, emailNotifications: e.target.checked })}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>

          <div className="pt-6 border-t border-gray-200">
            <button
              onClick={handleSave}
              className="btn-primary flex items-center gap-2"
            >
              <Save size={20} />
              Salvar Configurações
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
