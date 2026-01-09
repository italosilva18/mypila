import { useEffect, useCallback } from 'react';

/**
 * Hook to handle Escape key press for closing modals/dialogs
 * @param onEscape - Callback function to execute when Escape is pressed
 * @param isActive - Whether the hook should be active (e.g., when modal is open)
 */
export const useEscapeKey = (onEscape: () => void, isActive: boolean = true) => {
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        onEscape();
      }
    },
    [onEscape]
  );

  useEffect(() => {
    if (!isActive) return;

    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isActive, handleKeyDown]);
};
