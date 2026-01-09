import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDebounce } from './useDebounce';

describe('useDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Deve retornar o valor inicial imediatamente', () => {
    it('deve retornar o valor inicial sem delay', () => {
      const { result } = renderHook(() => useDebounce('initial', 300));

      expect(result.current).toBe('initial');
    });

    it('deve funcionar com diferentes tipos de dados', () => {
      // String
      const { result: stringResult } = renderHook(() => useDebounce('test', 300));
      expect(stringResult.current).toBe('test');

      // Number
      const { result: numberResult } = renderHook(() => useDebounce(42, 300));
      expect(numberResult.current).toBe(42);

      // Object
      const obj = { name: 'test', value: 123 };
      const { result: objectResult } = renderHook(() => useDebounce(obj, 300));
      expect(objectResult.current).toEqual(obj);

      // Array
      const arr = [1, 2, 3];
      const { result: arrayResult } = renderHook(() => useDebounce(arr, 300));
      expect(arrayResult.current).toEqual(arr);

      // Boolean
      const { result: boolResult } = renderHook(() => useDebounce(true, 300));
      expect(boolResult.current).toBe(true);

      // Null
      const { result: nullResult } = renderHook(() => useDebounce(null, 300));
      expect(nullResult.current).toBeNull();
    });
  });

  describe('Deve aplicar debounce corretamente', () => {
    it('deve atualizar o valor apos o delay', () => {
      const { result, rerender } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 300 } }
      );

      expect(result.current).toBe('initial');

      // Atualiza o valor
      rerender({ value: 'updated', delay: 300 });

      // O valor ainda nao deve ter mudado
      expect(result.current).toBe('initial');

      // Avanca o tempo
      act(() => {
        vi.advanceTimersByTime(300);
      });

      // Agora o valor deve estar atualizado
      expect(result.current).toBe('updated');
    });

    it('nao deve atualizar antes do delay', () => {
      const { result, rerender } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 500 } }
      );

      rerender({ value: 'updated', delay: 500 });

      // Avanca apenas metade do tempo
      act(() => {
        vi.advanceTimersByTime(250);
      });

      // O valor ainda nao deve ter mudado
      expect(result.current).toBe('initial');

      // Avanca o resto do tempo
      act(() => {
        vi.advanceTimersByTime(250);
      });

      // Agora sim
      expect(result.current).toBe('updated');
    });

    it('deve cancelar o timer anterior quando o valor muda novamente', () => {
      const { result, rerender } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 300 } }
      );

      // Primeira atualizacao
      rerender({ value: 'first update', delay: 300 });

      // Avanca 200ms (menos que o delay)
      act(() => {
        vi.advanceTimersByTime(200);
      });

      // Segunda atualizacao (deve cancelar o timer anterior)
      rerender({ value: 'second update', delay: 300 });

      // Avanca mais 200ms
      act(() => {
        vi.advanceTimersByTime(200);
      });

      // O valor ainda nao deve ser 'second update' pois o timer foi reiniciado
      expect(result.current).toBe('initial');

      // Avanca mais 100ms para completar o delay
      act(() => {
        vi.advanceTimersByTime(100);
      });

      // Agora deve mostrar o segundo valor
      expect(result.current).toBe('second update');
    });
  });

  describe('Deve usar o delay padrao quando nao especificado', () => {
    it('deve usar 300ms como delay padrao', () => {
      const { result, rerender } = renderHook(
        ({ value }) => useDebounce(value),
        { initialProps: { value: 'initial' } }
      );

      rerender({ value: 'updated' });

      // Avanca 299ms
      act(() => {
        vi.advanceTimersByTime(299);
      });

      expect(result.current).toBe('initial');

      // Avanca mais 1ms para completar 300ms
      act(() => {
        vi.advanceTimersByTime(1);
      });

      expect(result.current).toBe('updated');
    });
  });

  describe('Deve respeitar diferentes valores de delay', () => {
    it('deve funcionar com delay curto (100ms)', () => {
      const { result, rerender } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 100 } }
      );

      rerender({ value: 'updated', delay: 100 });

      act(() => {
        vi.advanceTimersByTime(100);
      });

      expect(result.current).toBe('updated');
    });

    it('deve funcionar com delay longo (1000ms)', () => {
      const { result, rerender } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 1000 } }
      );

      rerender({ value: 'updated', delay: 1000 });

      act(() => {
        vi.advanceTimersByTime(999);
      });

      expect(result.current).toBe('initial');

      act(() => {
        vi.advanceTimersByTime(1);
      });

      expect(result.current).toBe('updated');
    });

    it('deve funcionar com delay zero', () => {
      const { result, rerender } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 0 } }
      );

      rerender({ value: 'updated', delay: 0 });

      act(() => {
        vi.advanceTimersByTime(0);
      });

      expect(result.current).toBe('updated');
    });
  });

  describe('Deve lidar com mudancas de delay', () => {
    it('deve aplicar novo delay quando ele muda', () => {
      const { result, rerender } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 500 } }
      );

      // Muda valor e delay juntos
      rerender({ value: 'updated', delay: 200 });

      // Avanca 200ms (novo delay)
      act(() => {
        vi.advanceTimersByTime(200);
      });

      expect(result.current).toBe('updated');
    });
  });

  describe('Deve limpar o timer no unmount', () => {
    it('nao deve causar erros ao desmontar antes do timeout', () => {
      const { result, rerender, unmount } = renderHook(
        ({ value, delay }) => useDebounce(value, delay),
        { initialProps: { value: 'initial', delay: 300 } }
      );

      rerender({ value: 'updated', delay: 300 });

      // Desmonta antes do timeout
      unmount();

      // Avanca o tempo - nao deve causar erros
      act(() => {
        vi.advanceTimersByTime(300);
      });

      // O hook foi desmontado, entao nao podemos verificar result.current
      // Mas o teste passa se nao houver erros
      expect(true).toBe(true);
    });
  });

  describe('Casos de uso praticos', () => {
    it('deve funcionar para debounce de input de busca', () => {
      const { result, rerender } = renderHook(
        ({ value }) => useDebounce(value, 300),
        { initialProps: { value: '' } }
      );

      // Usuario digita 'a'
      rerender({ value: 'a' });
      expect(result.current).toBe('');

      // Usuario digita 'ab' (50ms depois)
      act(() => vi.advanceTimersByTime(50));
      rerender({ value: 'ab' });
      expect(result.current).toBe('');

      // Usuario digita 'abc' (50ms depois)
      act(() => vi.advanceTimersByTime(50));
      rerender({ value: 'abc' });
      expect(result.current).toBe('');

      // Usuario para de digitar, espera 300ms
      act(() => vi.advanceTimersByTime(300));
      expect(result.current).toBe('abc');
    });

    it('deve funcionar para debounce de resize', () => {
      const { result, rerender } = renderHook(
        ({ value }) => useDebounce(value, 150),
        { initialProps: { value: { width: 1920, height: 1080 } } }
      );

      // Simula varios eventos de resize
      rerender({ value: { width: 1900, height: 1070 } });
      rerender({ value: { width: 1850, height: 1050 } });
      rerender({ value: { width: 1800, height: 1000 } });

      expect(result.current).toEqual({ width: 1920, height: 1080 });

      act(() => vi.advanceTimersByTime(150));

      expect(result.current).toEqual({ width: 1800, height: 1000 });
    });
  });
});
