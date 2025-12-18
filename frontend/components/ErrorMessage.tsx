import React from 'react';
import { AlertCircle } from 'lucide-react';

interface Props {
  error?: string;
  className?: string;
}

export const ErrorMessage: React.FC<Props> = ({ error, className = '' }) => {
  if (!error) return null;

  return (
    <div className={`flex items-start gap-1.5 text-red-500 text-xs mt-1 ${className}`}>
      <AlertCircle className="w-3 h-3 mt-0.5 flex-shrink-0" />
      <span>{error}</span>
    </div>
  );
};
