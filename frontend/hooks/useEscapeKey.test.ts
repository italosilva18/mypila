import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useEscapeKey } from './useEscapeKey';

describe('useEscapeKey', () => {
  let addEventListenerSpy: ReturnType<typeof vi.spyOn>;
  let removeEventListenerSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    addEventListenerSpy = vi.spyOn(document, 'addEventListener');
    removeEventListenerSpy = vi.spyOn(document, 'removeEventListener');
  });

  afterEach(() => {
    addEventListenerSpy.mockRestore();
    removeEventListenerSpy.mockRestore();
  });

  describe('Deve registrar event listener quando ativo', () => {
    it('deve adicionar event listener ao montar quando isActive e true', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, true));

      expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    });

    it('deve adicionar event listener com isActive padrao (true)', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape));

      expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    });

    it('nao deve adicionar event listener quando isActive e false', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, false));

      expect(addEventListenerSpy).not.toHaveBeenCalled();
    });
  });

  describe('Deve remover event listener ao desmontar', () => {
    it('deve remover event listener ao desmontar', () => {
      const onEscape = vi.fn();

      const { unmount } = renderHook(() => useEscapeKey(onEscape, true));

      unmount();

      expect(removeEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    });

    it('nao deve chamar removeEventListener se nunca foi adicionado', () => {
      const onEscape = vi.fn();

      const { unmount } = renderHook(() => useEscapeKey(onEscape, false));

      unmount();

      expect(removeEventListenerSpy).not.toHaveBeenCalled();
    });
  });

  describe('Deve chamar callback quando Escape e pressionado', () => {
    it('deve chamar onEscape quando tecla Escape e pressionada', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, true));

      // Simula o evento de tecla Escape
      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      expect(onEscape).toHaveBeenCalledTimes(1);
    });

    it('deve prevenir comportamento padrao do evento', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, true));

      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      const preventDefaultSpy = vi.spyOn(escapeEvent, 'preventDefault');

      document.dispatchEvent(escapeEvent);

      expect(preventDefaultSpy).toHaveBeenCalled();
    });

    it('nao deve chamar onEscape quando outras teclas sao pressionadas', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, true));

      // Simula outras teclas
      const enterEvent = new KeyboardEvent('keydown', { key: 'Enter' });
      document.dispatchEvent(enterEvent);

      const spaceEvent = new KeyboardEvent('keydown', { key: ' ' });
      document.dispatchEvent(spaceEvent);

      const aEvent = new KeyboardEvent('keydown', { key: 'a' });
      document.dispatchEvent(aEvent);

      expect(onEscape).not.toHaveBeenCalled();
    });

    it('nao deve chamar onEscape quando isActive e false', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, false));

      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      expect(onEscape).not.toHaveBeenCalled();
    });
  });

  describe('Deve lidar com mudancas de isActive', () => {
    it('deve adicionar listener quando isActive muda de false para true', () => {
      const onEscape = vi.fn();

      const { rerender } = renderHook(
        ({ isActive }) => useEscapeKey(onEscape, isActive),
        { initialProps: { isActive: false } }
      );

      expect(addEventListenerSpy).not.toHaveBeenCalled();

      rerender({ isActive: true });

      expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    });

    it('deve remover listener quando isActive muda de true para false', () => {
      const onEscape = vi.fn();

      const { rerender } = renderHook(
        ({ isActive }) => useEscapeKey(onEscape, isActive),
        { initialProps: { isActive: true } }
      );

      expect(addEventListenerSpy).toHaveBeenCalled();

      rerender({ isActive: false });

      expect(removeEventListenerSpy).toHaveBeenCalled();
    });

    it('deve responder ao Escape apos reativar', () => {
      const onEscape = vi.fn();

      const { rerender } = renderHook(
        ({ isActive }) => useEscapeKey(onEscape, isActive),
        { initialProps: { isActive: true } }
      );

      // Desativa
      rerender({ isActive: false });

      // Escape nao deve funcionar
      let escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);
      expect(onEscape).not.toHaveBeenCalled();

      // Reativa
      rerender({ isActive: true });

      // Agora Escape deve funcionar
      escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);
      expect(onEscape).toHaveBeenCalledTimes(1);
    });
  });

  describe('Deve lidar com mudancas de callback', () => {
    it('deve usar o callback atualizado', () => {
      const onEscape1 = vi.fn();
      const onEscape2 = vi.fn();

      const { rerender } = renderHook(
        ({ onEscape }) => useEscapeKey(onEscape, true),
        { initialProps: { onEscape: onEscape1 } }
      );

      // Atualiza o callback
      rerender({ onEscape: onEscape2 });

      // Dispara Escape
      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      // Deve chamar o novo callback
      expect(onEscape2).toHaveBeenCalledTimes(1);
      // O callback antigo nao deve ser chamado
      expect(onEscape1).not.toHaveBeenCalled();
    });
  });

  describe('Casos de uso praticos', () => {
    it('deve funcionar para fechar modal', () => {
      const closeModal = vi.fn();
      const isModalOpen = true;

      renderHook(() => useEscapeKey(closeModal, isModalOpen));

      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      expect(closeModal).toHaveBeenCalled();
    });

    it('deve ignorar Escape quando modal esta fechado', () => {
      const closeModal = vi.fn();
      const isModalOpen = false;

      renderHook(() => useEscapeKey(closeModal, isModalOpen));

      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      expect(closeModal).not.toHaveBeenCalled();
    });

    it('deve funcionar com multiplos hooks ativos', () => {
      const closeModal1 = vi.fn();
      const closeModal2 = vi.fn();

      renderHook(() => useEscapeKey(closeModal1, true));
      renderHook(() => useEscapeKey(closeModal2, true));

      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      // Ambos devem ser chamados
      expect(closeModal1).toHaveBeenCalled();
      expect(closeModal2).toHaveBeenCalled();
    });

    it('deve funcionar para cancelar edicao', () => {
      let isEditing = true;
      const cancelEdit = vi.fn(() => {
        isEditing = false;
      });

      renderHook(() => useEscapeKey(cancelEdit, isEditing));

      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      expect(cancelEdit).toHaveBeenCalled();
      expect(isEditing).toBe(false);
    });
  });

  describe('Variantes da tecla Escape', () => {
    it('deve responder apenas a key "Escape"', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, true));

      // A tecla padrao e "Escape"
      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      expect(onEscape).toHaveBeenCalledTimes(1);
    });

    it('nao deve responder a "Esc" (formato antigo)', () => {
      const onEscape = vi.fn();

      renderHook(() => useEscapeKey(onEscape, true));

      // Formato antigo do IE
      const escEvent = new KeyboardEvent('keydown', { key: 'Esc' });
      document.dispatchEvent(escEvent);

      // A implementacao atual usa apenas 'Escape'
      expect(onEscape).not.toHaveBeenCalled();
    });
  });

  describe('Cleanup e memoria', () => {
    it('deve fazer cleanup correto em multiplos rerenders', () => {
      const onEscape = vi.fn();

      const { rerender, unmount } = renderHook(
        ({ isActive }) => useEscapeKey(onEscape, isActive),
        { initialProps: { isActive: true } }
      );

      // Varios rerenders
      rerender({ isActive: true });
      rerender({ isActive: true });
      rerender({ isActive: true });

      // Apenas um listener deve estar ativo
      const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape' });
      document.dispatchEvent(escapeEvent);

      expect(onEscape).toHaveBeenCalledTimes(1);

      unmount();

      // Apos unmount, novo escape nao deve chamar o callback
      onEscape.mockClear();
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
      expect(onEscape).not.toHaveBeenCalled();
    });
  });
});
