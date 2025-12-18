/**
 * EXEMPLO DE USO DO SISTEMA DE TOAST
 *
 * Este arquivo demonstra como utilizar o sistema de notificacoes toast
 * em diferentes cenarios dentro da aplicacao.
 */

import React from 'react';
import { useToast } from '../contexts/ToastContext';

export const ToastExample: React.FC = () => {
  const { addToast } = useToast();

  const handleSuccess = () => {
    addToast('success', 'Operacao realizada com sucesso!');
  };

  const handleError = () => {
    addToast('error', 'Ocorreu um erro ao processar sua solicitacao.');
  };

  const handleWarning = () => {
    addToast('warning', 'Atencao: Esta acao pode ter consequencias.');
  };

  const handleInfo = () => {
    addToast('info', 'Informacao importante sobre o sistema.');
  };

  const handleMultiple = () => {
    addToast('info', 'Iniciando processo...');
    setTimeout(() => {
      addToast('warning', 'Processando dados...');
    }, 1000);
    setTimeout(() => {
      addToast('success', 'Processo concluido com sucesso!');
    }, 2000);
  };

  return (
    <div className="p-8 space-y-4">
      <h2 className="text-2xl font-bold text-ink mb-6">
        Exemplos de Notificacoes Toast
      </h2>

      <div className="grid grid-cols-2 gap-4 max-w-2xl">
        <button
          onClick={handleSuccess}
          className="px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
        >
          Exibir Toast de Sucesso
        </button>

        <button
          onClick={handleError}
          className="px-4 py-2 bg-rose-600 text-white rounded-lg hover:bg-rose-700 transition-colors"
        >
          Exibir Toast de Erro
        </button>

        <button
          onClick={handleWarning}
          className="px-4 py-2 bg-amber-600 text-white rounded-lg hover:bg-amber-700 transition-colors"
        >
          Exibir Toast de Aviso
        </button>

        <button
          onClick={handleInfo}
          className="px-4 py-2 bg-sky-600 text-white rounded-lg hover:bg-sky-700 transition-colors"
        >
          Exibir Toast de Info
        </button>

        <button
          onClick={handleMultiple}
          className="col-span-2 px-4 py-2 bg-stone-700 text-white rounded-lg hover:bg-stone-800 transition-colors"
        >
          Exibir Multiplos Toasts
        </button>
      </div>

      <div className="mt-8 p-4 bg-paper-dark rounded-xl">
        <h3 className="text-lg font-semibold text-ink mb-2">Como usar em componentes:</h3>
        <pre className="text-sm text-ink-muted bg-white/50 p-4 rounded-lg overflow-x-auto">
{`import { useToast } from '../contexts/ToastContext';

const MyComponent = () => {
  const { addToast } = useToast();

  const handleAction = async () => {
    try {
      // Sua logica aqui
      await someApiCall();
      addToast('success', 'Acao concluida!');
    } catch (error) {
      addToast('error', 'Erro ao executar acao');
    }
  };

  return <button onClick={handleAction}>Executar</button>;
};`}
        </pre>
      </div>

      <div className="mt-6 p-4 bg-amber-50 border-2 border-amber-200 rounded-xl">
        <h3 className="text-lg font-semibold text-amber-900 mb-2">Tipos de Toast:</h3>
        <ul className="list-disc list-inside text-amber-900 space-y-1">
          <li><strong>success:</strong> Operacoes concluidas com sucesso (verde)</li>
          <li><strong>error:</strong> Erros e falhas (vermelho)</li>
          <li><strong>warning:</strong> Avisos e alertas (amarelo)</li>
          <li><strong>info:</strong> Informacoes gerais (azul)</li>
        </ul>
      </div>

      <div className="mt-6 p-4 bg-sky-50 border-2 border-sky-200 rounded-xl">
        <h3 className="text-lg font-semibold text-sky-900 mb-2">Caracteristicas:</h3>
        <ul className="list-disc list-inside text-sky-900 space-y-1">
          <li>Auto-dismiss apos 5 segundos</li>
          <li>Botao de fechar manual</li>
          <li>Animacao de entrada e saida</li>
          <li>Empilhamento vertical</li>
          <li>Tema vintage/stone consistente</li>
        </ul>
      </div>
    </div>
  );
};
