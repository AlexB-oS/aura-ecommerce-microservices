// Shared TypeScript types that mirror the user service's JSON contract.

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number;
}

// register/login return tokens plus the user; refresh returns tokens only.
export interface AuthResponse extends AuthTokens {
  user?: User;
}

// The error shape inside { "error": { ... } } from every service.
export interface ApiErrorBody {
  code: string;
  message: string;
  fields?: Record<string, string>;
}

// --- Orders ----------------------------------------------------------------

export interface OrderItem {
  productId: string;
  slug: string;
  name: string;
  priceCents: number;
  qty: number;
}

export interface Order {
  id: string;
  status: "pending_payment" | "paid" | "payment_failed" | "cancelled";
  totalCents: number;
  currency: string;
  items?: OrderItem[];
  createdAt: string;
  updatedAt: string;
}
