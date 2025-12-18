# Guia de Estilo - Sistema de Validação

## Padrões Visuais

### 1. Estados dos Campos de Input

#### Campo Normal (Sem interação)
```jsx
<input
  className="w-full px-4 py-2.5 bg-stone-50 border border-stone-200 rounded-xl
             text-stone-900 placeholder-stone-400 focus:outline-none
             focus:ring-2 focus:ring-stone-400 transition-colors"
/>
```

**Características:**
- Fundo: `bg-stone-50`
- Borda: `border-stone-200`
- Texto: `text-stone-900`
- Placeholder: `placeholder-stone-400`
- Focus ring: `focus:ring-stone-400`

#### Campo com Erro
```jsx
<input
  className="w-full px-4 py-2.5 bg-stone-50 border border-red-500 rounded-xl
             text-stone-900 placeholder-stone-400 focus:outline-none
             focus:ring-2 focus:ring-red-400 transition-colors"
/>
```

**Características:**
- Borda: `border-red-500` (vermelho)
- Focus ring: `focus:ring-red-400` (vermelho mais claro)
- Transição suave entre estados

### 2. Labels

#### Label Normal
```jsx
<label className="block text-sm font-medium text-stone-600 mb-1.5">
  Nome do Campo
</label>
```

#### Label com Campo Obrigatório
```jsx
<label className="block text-sm font-medium text-stone-600 mb-1.5">
  Nome do Campo <span className="text-red-500">*</span>
</label>
```

**Características:**
- Asterisco vermelho para indicar obrigatoriedade
- Tamanho: `text-sm`
- Peso: `font-medium`
- Cor: `text-stone-600`
- Margem inferior: `mb-1.5`

### 3. Mensagens de Erro

#### Estrutura
```jsx
<div className="flex items-start gap-1.5 text-red-500 text-xs mt-1">
  <AlertCircle className="w-3 h-3 mt-0.5 flex-shrink-0" />
  <span>Mensagem de erro aqui</span>
</div>
```

**Características:**
- Cor: `text-red-500`
- Tamanho: `text-xs`
- Margem superior: `mt-1`
- Ícone: AlertCircle do lucide-react
- Alinhamento: flex com gap pequeno

#### Usando o Componente ErrorMessage
```jsx
<ErrorMessage error={getError('fieldName')} />
```

### 4. Textos de Ajuda

```jsx
<p className="text-xs text-stone-400 mt-1">
  Texto de ajuda ou informação adicional
</p>
```

**Características:**
- Cor: `text-stone-400` (mais clara que erro)
- Tamanho: `text-xs`
- Margem superior: `mt-1`

### 5. Botões

#### Botão Normal (Ativo)
```jsx
<button
  type="submit"
  className="w-full bg-stone-800 hover:bg-stone-700 text-white font-medium
             py-3 rounded-xl transition-all shadow-lg shadow-stone-900/20
             active:scale-[0.98] flex items-center justify-center gap-2"
>
  <Save className="w-4 h-4" />
  Salvar
</button>
```

#### Botão Desabilitado (Com Erros)
```jsx
<button
  type="submit"
  disabled={true}
  className="w-full bg-stone-300 text-stone-500 cursor-not-allowed
             font-medium py-3 rounded-xl transition-all shadow-lg
             flex items-center justify-center gap-2"
>
  <Save className="w-4 h-4" />
  Salvar
</button>
```

**Características do Estado Desabilitado:**
- Fundo: `bg-stone-300` (cinza claro)
- Texto: `text-stone-500` (cinza médio)
- Cursor: `cursor-not-allowed`
- Sem hover effects
- Sem active scale

#### Botão com Estado Condicional
```jsx
<button
  type="submit"
  disabled={hasErrors()}
  className={`w-full font-medium py-3 rounded-xl transition-all shadow-lg
              flex items-center justify-center gap-2 ${
    hasErrors()
      ? 'bg-stone-300 text-stone-500 cursor-not-allowed'
      : 'bg-stone-800 hover:bg-stone-700 text-white shadow-stone-900/20 active:scale-[0.98]'
  }`}
>
  <Save className="w-4 h-4" />
  Salvar
</button>
```

### 6. Containers de Formulário

#### Form Container Padrão
```jsx
<form className="p-6 space-y-4">
  {/* Campos aqui */}
</form>
```

#### Form Container com Borda
```jsx
<form className="bg-white border border-stone-200 rounded-2xl p-6 space-y-4">
  {/* Campos aqui */}
</form>
```

**Características:**
- Espaçamento: `space-y-4` (entre campos)
- Padding: `p-6`
- Background: `bg-white`
- Borda: `border border-stone-200`
- Arredondamento: `rounded-2xl`

### 7. Grid Layouts

#### Duas Colunas
```jsx
<div className="grid grid-cols-2 gap-4">
  <div>{/* Campo 1 */}</div>
  <div>{/* Campo 2 */}</div>
</div>
```

#### Responsivo
```jsx
<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
  <div>{/* Campo 1 */}</div>
  <div>{/* Campo 2 */}</div>
</div>
```

#### Campo que Ocupa Toda a Largura
```jsx
<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
  <div className="md:col-span-2">
    {/* Campo que ocupa 2 colunas em desktop */}
  </div>
  <div>{/* Campo normal */}</div>
  <div>{/* Campo normal */}</div>
</div>
```

## Paleta de Cores

### Cores Principais
- **Stone 50**: `#fafaf9` - Fundos de inputs
- **Stone 200**: `#e7e5e4` - Bordas normais
- **Stone 300**: `#d6d3d1` - Botões desabilitados
- **Stone 400**: `#a8a29e` - Placeholders e textos de ajuda
- **Stone 500**: `#78716c` - Textos de botões desabilitados
- **Stone 600**: `#57534e` - Labels
- **Stone 700**: `#44403c` - Hover de botões
- **Stone 800**: `#292524` - Botões ativos
- **Stone 900**: `#1c1917` - Textos principais

### Cores de Estado
- **Red 400**: `#f87171` - Focus ring de erros
- **Red 500**: `#ef4444` - Bordas de erro e textos

### Cores de Sucesso (Opcional)
- **Green 400**: `#4ade80` - Focus ring de sucesso
- **Green 500**: `#22c55e` - Bordas de sucesso
- **Green 600**: `#16a34a` - Botões de ação positiva

## Exemplos Completos

### Campo de Texto Completo
```jsx
<div>
  <label className="block text-sm font-medium text-stone-600 mb-1.5">
    Descrição <span className="text-red-500">*</span>
  </label>
  <input
    type="text"
    value={formData.description}
    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
    placeholder="Digite a descrição..."
    className={`w-full px-4 py-2.5 bg-stone-50 border ${
      hasError('description')
        ? 'border-red-500 focus:ring-red-400'
        : 'border-stone-200 focus:ring-stone-400'
    } rounded-xl text-stone-900 placeholder-stone-400 focus:outline-none
    focus:ring-2 transition-colors`}
  />
  <ErrorMessage error={getError('description')} />
  <p className="text-xs text-stone-400 mt-1">Máximo 200 caracteres</p>
</div>
```

### Campo Numérico com Prefixo
```jsx
<div>
  <label className="block text-sm font-medium text-stone-600 mb-1.5">
    Valor <span className="text-red-500">*</span>
  </label>
  <div className="relative">
    <span className="absolute left-3 top-1/2 -translate-y-1/2 text-stone-400">
      R$
    </span>
    <input
      type="number"
      step="0.01"
      value={formData.amount || ''}
      onChange={(e) => setFormData({ ...formData, amount: parseFloat(e.target.value) || 0 })}
      placeholder="0,00"
      className={`w-full pl-10 pr-4 py-2.5 bg-stone-50 border ${
        hasError('amount')
          ? 'border-red-500 focus:ring-red-400'
          : 'border-stone-200 focus:ring-stone-400'
      } rounded-xl text-stone-900 placeholder-stone-400 focus:outline-none
      focus:ring-2 transition-colors`}
    />
  </div>
  <ErrorMessage error={getError('amount')} />
</div>
```

### Select/Dropdown
```jsx
<div>
  <label className="block text-sm font-medium text-stone-600 mb-1.5">
    Categoria <span className="text-red-500">*</span>
  </label>
  <select
    value={formData.category}
    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
    className={`w-full px-4 py-2.5 bg-stone-50 border ${
      hasError('category')
        ? 'border-red-500 focus:ring-red-400'
        : 'border-stone-200 focus:ring-stone-400'
    } rounded-xl text-stone-900 focus:outline-none focus:ring-2 transition-colors`}
  >
    <option value="">Selecione...</option>
    <option value="receita">Receita</option>
    <option value="despesa">Despesa</option>
  </select>
  <ErrorMessage error={getError('category')} />
</div>
```

### Textarea
```jsx
<div>
  <label className="block text-sm font-medium text-stone-600 mb-1.5">
    Observações
  </label>
  <textarea
    value={formData.notes}
    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
    placeholder="Digite observações..."
    rows={3}
    className={`w-full px-4 py-2.5 bg-stone-50 border ${
      hasError('notes')
        ? 'border-red-500 focus:ring-red-400'
        : 'border-stone-200 focus:ring-stone-400'
    } rounded-xl text-stone-900 placeholder-stone-400 focus:outline-none
    focus:ring-2 transition-colors resize-none`}
  />
  <ErrorMessage error={getError('notes')} />
  <p className="text-xs text-stone-400 mt-1">
    {formData.notes.length}/500 caracteres
  </p>
</div>
```

## Animações e Transições

### Transições Padrão
```css
transition-colors /* Para mudanças de cor */
transition-all    /* Para múltiplas propriedades */
```

### Hover Effects
```jsx
hover:bg-stone-700     /* Botões */
hover:border-stone-300 /* Inputs e containers */
hover:text-stone-800   /* Textos */
```

### Active Effects
```jsx
active:scale-[0.98]  /* Botões - feedback tátil */
```

### Focus Effects
```jsx
focus:outline-none
focus:ring-2
focus:ring-stone-400  /* ou focus:ring-red-400 para erros */
```

## Acessibilidade

### Checklist
- ✅ Todas as labels devem estar associadas aos inputs
- ✅ Campos obrigatórios devem ter asterisco visível
- ✅ Mensagens de erro devem ser descritivas
- ✅ Contraste adequado entre texto e fundo (WCAG AA)
- ✅ Focus visível em todos os elementos interativos
- ✅ Estados disabled claramente indicados

### ARIA (Opcional, para melhorar ainda mais)
```jsx
<input
  aria-required="true"
  aria-invalid={hasError('fieldName')}
  aria-describedby={hasError('fieldName') ? 'fieldName-error' : undefined}
/>
{hasError('fieldName') && (
  <div id="fieldName-error" role="alert">
    <ErrorMessage error={getError('fieldName')} />
  </div>
)}
```

## Responsividade

### Breakpoints
- `sm:` - 640px
- `md:` - 768px
- `lg:` - 1024px
- `xl:` - 1280px

### Padrões Mobile-First
```jsx
{/* Mobile: 1 coluna, Desktop: 2 colunas */}
<div className="grid grid-cols-1 md:grid-cols-2 gap-4">

{/* Mobile: stack vertical, Desktop: horizontal */}
<div className="flex flex-col sm:flex-row gap-4">

{/* Padding responsivo */}
<div className="p-4 sm:p-6 lg:p-8">
```

## Boas Práticas

1. **Consistência**: Use sempre as mesmas classes para os mesmos elementos
2. **Feedback Visual**: Sempre indique quando um campo tem erro
3. **Transições**: Use `transition-colors` para mudanças suaves
4. **Espaçamento**: Mantenha `space-y-4` entre campos
5. **Mensagens Claras**: Erros devem ser específicos e acionáveis
6. **Desabilitar Submit**: Sempre desabilite o botão quando há erros
7. **Limpar Erros**: Limpe erros ao fechar modais/forms

## Ferramentas

### Extensões VS Code Recomendadas
- Tailwind CSS IntelliSense
- PostCSS Language Support
- Headwind (ordena classes automaticamente)

### Recursos
- [Tailwind CSS Docs](https://tailwindcss.com/docs)
- [Lucide React Icons](https://lucide.dev)
- [WCAG Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
