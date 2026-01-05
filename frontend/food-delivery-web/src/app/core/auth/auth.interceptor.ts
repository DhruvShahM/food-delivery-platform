import { HttpInterceptorFn, HttpRequest } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from './auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const token = auth.getToken();

  // Do not attach token for auth endpoints
  const isAuthEndpoint = /\/auth\/(login|register|logout)/.test(req.url);
  if (!token || isAuthEndpoint) {
    return next(req);
  }

  const authReq: HttpRequest<unknown> = req.clone({
    setHeaders: {
      Authorization: `Bearer ${token}`,
    },
  });
  return next(authReq);
};
