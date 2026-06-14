package com.generate.demo.common.web;

import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.http.HttpServletRequest;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.stereotype.Component;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;

import java.util.Arrays;

/**
 * API 请求日志切面
 * <p>
 * 自动打印所有 RestController 的输入参数和返回结果。
 * 异常时打印完整堆栈且不吞异常。
 */
@Slf4j
@Aspect
@Component
@RequiredArgsConstructor
public class LoggingAspect {

    private final ObjectMapper objectMapper;

    @Around("execution(* com.generate.demo..*Controller.*(..))")
    public Object logRequestResponse(ProceedingJoinPoint joinPoint) throws Throwable {
        // 请求信息
        String className = joinPoint.getTarget().getClass().getSimpleName();
        String methodName = joinPoint.getSignature().getName();
        Object[] args = joinPoint.getArgs();

        // 过滤掉 HttpServletRequest/Response 等无法序列化的参数
        String argsStr = maskArgs(args);

        // 获取请求路径
        String path = "";
        ServletRequestAttributes attrs = (ServletRequestAttributes) RequestContextHolder.getRequestAttributes();
        if (attrs != null) {
            HttpServletRequest request = attrs.getRequest();
            path = request.getMethod() + " " + request.getRequestURI();
        }

        log.info("→ {}.{}() | {} | args={}", className, methodName, path, argsStr);

        long start = System.currentTimeMillis();
        try {
            Object result = joinPoint.proceed();
            long elapsed = System.currentTimeMillis() - start;
            log.info("← {}.{}() | {}ms | result={}", className, methodName, elapsed,
                    safeToJson(result));
            return result;
        } catch (Exception e) {
            long elapsed = System.currentTimeMillis() - start;
            log.error("✕ {}.{}() | {}ms | error={}", className, methodName, elapsed,
                    e.getMessage(), e);  // 打印堆栈，不吞异常
            throw e;
        }
    }

    private String maskArgs(Object[] args) {
        if (args == null || args.length == 0) return "[]";
        // 过滤掉框架内部参数
        Object[] filtered = Arrays.stream(args)
                .filter(a -> a != null && !isFrameworkType(a.getClass()))
                .toArray();
        if (filtered.length == 0) return "[]";
        return safeToJson(filtered);
    }

    private boolean isFrameworkType(Class<?> clazz) {
        return clazz.getName().startsWith("jakarta.servlet")
                || clazz.getName().startsWith("org.springframework");
    }

    private String safeToJson(Object obj) {
        try {
            return objectMapper.writeValueAsString(obj);
        } catch (Exception e) {
            return String.valueOf(obj);
        }
    }
}
