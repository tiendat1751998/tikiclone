declare module 'next/font/google' {
  export interface FontOptions {
    subsets: string[];
    display?: 'auto' | 'block' | 'swap' | 'fallback' | 'optional';
    weight?: string | string[];
    style?: string | string[];
    variable?: string;
  }
  export function Inter(options: FontOptions): {
    className: string;
    style: { fontFamily: string };
  };
}

/// <reference types="next" />
/// <reference types="next/image-types/global" />
