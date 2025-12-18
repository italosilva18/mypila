# Sistema de Notificacoes Toast

Sistema elegante de notificacoes toast implementado com React Context API, seguindo o tema visual vintage/stone do projeto.

## Arquitetura

### Arquivos Criados

```
frontend/
├── contexts/
│   └── ToastContext.tsx          # Contexto e Provider do sistema de toast
├── components/
│   └── ToastContainer.tsx        # Componente de renderizacao dos toasts
└── examples/
    └── ToastExample.tsx          # Exemplos de uso
```

## Como Usar

### 1. Setup Basico (Ja Integrado)

O sistema ja esta integrado no `App.tsx`:

```tsx
import { ToastProvider } from './contexts/ToastContext';
import { ToastContainer } from './components/ToastContainer';

<ToastProvider>
  <AppRoutes />
  <ToastContainer />
</ToastProvider>
```

### 2. Usando em Componentes

```tsx
import { useToast } from '../contexts/ToastContext';

const MyComponent = () => {
  const { addToast } = useToast();

  const handleSuccess = () => {
    addToast('success', 'Operacao concluida com sucesso!');
  };

  const handleError = () => {
    addToast('error', 'Erro ao processar a solicitacao');
  };

  return (
    <button onClick={handleSuccess}>Executar Acao</button>
  );
};
```

### 3. Tipos de Toast

#### Success (Verde)
```tsx
addToast('success', 'Transacao criada com sucesso!');
```
- Usado para: Operacoes concluidas, saves, criacao de registros
- Cor: Emerald (verde)
- Icone: CheckCircle2

#### Error (Vermelho)
```tsx
addToast('error', 'Erro ao conectar com o servidor');
```
- Usado para: Erros, falhas, problemas
- Cor: Rose (vermelho)
- Icone: XCircle

#### Warning (Amarelo)
```tsx
addToast('warning', 'Esta acao nao pode ser desfeita');
```
- Usado para: Avisos, alertas, atencao necessaria
- Cor: Amber (amarelo)
- Icone: AlertTriangle

#### Info (Azul)
```tsx
addToast('info', 'Status alterado para pago');
```
- Usado para: Informacoes gerais, mudancas de estado
- Cor: Sky (azul)
- Icone: Info

## Caracteristicas

### Auto-Dismiss
- Toasts desaparecem automaticamente apos 5 segundos
- Implementado com setTimeout no contexto

### Animacoes
- **Entrada**: Slide da direita com fade-in
- **Saida**: Slide para direita com fade-out
- Duracao: 300ms com easing ease-out

### Interatividade
- Botao X para fechar manualmente
- Hover states nos botoes
- Acessibilidade com aria-labels

### Design
- Tema vintage/stone consistente com o projeto
- Backdrop blur para efeito glassmorphism
- Bordas arredondadas (rounded-xl)
- Sombras suaves (shadow-card)
- Cores semanticas por tipo

### Responsividade
- Posicionamento fixo no canto superior direito
- Min-width: 320px, Max-width: 450px
- Empilhamento vertical com gap de 12px
- z-index: 50 para ficar acima de outros elementos

## Integracao com useTransactions

O hook `useTransactions` foi atualizado para usar toasts:

```typescript
// Sucesso ao criar transacao
addToast('success', 'Transacao criada com sucesso!');

// Erro ao criar transacao
addToast('error', `Erro ao criar transacao: ${message}`);

// Info ao mudar status
addToast('info', `Status alterado para ${status}`);
```

### Operacoes com Toast

- **createTransaction**: Toast de sucesso ou erro
- **updateTransaction**: Toast de sucesso ou erro
- **deleteTransaction**: Toast de sucesso ou erro
- **toggleStatus**: Toast info com novo status

## API do Context

### ToastProvider Props
```typescript
interface ToastProviderProps {
  children: ReactNode;
}
```

### useToast Hook
```typescript
const { toasts, addToast, removeToast } = useToast();

// toasts: Toast[] - Lista atual de toasts
// addToast: (type, message) => void - Adiciona novo toast
// removeToast: (id) => void - Remove toast especifico
```

### Toast Interface
```typescript
interface Toast {
  id: string;              // ID unico gerado automaticamente
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;         // Texto da notificacao
}
```

## Boas Praticas

### 1. Mensagens Claras
```typescript
// BOM
addToast('success', 'Transacao #1234 criada com sucesso!');

// EVITE
addToast('success', 'OK');
```

### 2. Contexto Adequado
```typescript
// BOM - Mensagem especifica
addToast('error', 'Erro ao conectar: timeout de rede');

// EVITE - Mensagem generica
addToast('error', 'Erro');
```

### 3. Tipo Correto
```typescript
// BOM - Info para mudancas de estado
addToast('info', 'Status alterado para pago');

// EVITE - Success para simples informacao
addToast('success', 'Status alterado');
```

### 4. Nao Abusar
```typescript
// BOM - Toast apenas para acoes importantes
const handleSave = async () => {
  await saveData();
  addToast('success', 'Dados salvos!');
};

// EVITE - Toast para cada interacao pequena
const handleClick = () => {
  addToast('info', 'Botao clicado');
};
```

## Acessibilidade

### Atributos ARIA
```tsx
<div aria-live="polite" aria-atomic="true">
  {/* Toasts aqui */}
</div>
```

- `aria-live="polite"`: Anuncia mudancas sem interromper
- `aria-atomic="true"`: Le o container completo
- `aria-label` nos botoes de fechar

### Teclado
- Botao X e focavel e acessivel via Tab
- Enter/Space para fechar toast

## Performance

### Otimizacoes
- useCallback para prevenir re-renders
- setTimeout limpo automaticamente
- Remocao eficiente com filter
- Animacoes com CSS transforms (GPU-accelerated)

### Geracao de IDs
```typescript
const id = `toast-${Date.now()}-${++toastCounter}`;
```
- Combinacao de timestamp + contador
- Garante IDs unicos mesmo em rapid fire

## Customizacao Futura

### Adicionar Duracao Customizada
```typescript
addToast('success', 'Mensagem', { duration: 3000 });
```

### Adicionar Acoes
```typescript
addToast('warning', 'Deseja continuar?', {
  action: { label: 'Desfazer', onClick: handleUndo }
});
```

### Posicionamento Customizado
```typescript
<ToastContainer position="bottom-right" />
```

## Troubleshooting

### Toast nao aparece
- Verifique se o componente esta dentro do ToastProvider
- Verifique z-index de outros elementos
- Verifique console para erros

### Animacao nao funciona
- Verifique se Tailwind esta carregado
- Verifique se as classes de transicao existem
- Teste em navegador diferente

### Multiple toasts sobrepostos
- Isso e esperado - design empilha verticalmente
- Para limitar, implemente max toasts no contexto

## Exemplo Completo

Veja `examples/ToastExample.tsx` para exemplos interativos de todos os tipos de toast.

## Tecnologias

- React 18+
- TypeScript
- Tailwind CSS
- Lucide React (icones)
- React Context API
