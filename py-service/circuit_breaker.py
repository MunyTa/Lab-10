import time
import logging
from functools import wraps
from typing import Callable, Any

logger = logging.getLogger(__name__)

class CircuitBreaker:
    """Circuit Breaker для защиты от падающего Go сервиса"""
    
    def __init__(self, failure_threshold: int = 5, recovery_timeout: int = 30):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.failure_count = 0
        self.last_failure_time = 0
        self.state = "CLOSED"  # CLOSED, OPEN, HALF_OPEN
        
    async def call_async(self, func: Callable, *args, **kwargs) -> Any:
        """Асинхронный вызов функции с защитой Circuit Breaker"""
        
        if self.state == "OPEN":
            if time.time() - self.last_failure_time > self.recovery_timeout:
                logger.info("Circuit Breaker переходит в HALF_OPEN состояние")
                self.state = "HALF_OPEN"
            else:
                raise Exception("Circuit Breaker OPEN - сервис временно недоступен")
        
        try:
            result = await func(*args, **kwargs)
            
            if self.state == "HALF_OPEN":
                logger.info("Circuit Breaker переходит в CLOSED состояние")
                self.state = "CLOSED"
                self.failure_count = 0
                
            return result
            
        except Exception as e:
            self.failure_count += 1
            self.last_failure_time = time.time()
            
            if self.failure_count >= self.failure_threshold:
                self.state = "OPEN"
                logger.error(f"Circuit Breaker OPEN после {self.failure_count} ошибок")
                
            raise e

# Глобальный экземпляр для Go сервиса
go_circuit_breaker = CircuitBreaker(failure_threshold=3, recovery_timeout=30)