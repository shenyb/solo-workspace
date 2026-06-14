package {{basePackage}}.common.web;

import jakarta.servlet.*;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.MDC;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.UUID;

/**
 * 全链路 traceId 过滤器
 * <p>
 * 每次请求生成 traceId，写入 MDC 和响应头 X-Trace-Id。
 * 日志配置 %X{traceId} 即可自动输出 traceId。
 */
@Slf4j
@Component
@Order(1)
public class MdcFilter implements Filter {

    private static final String TRACE_ID = "traceId";

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {

        String traceId = UUID.randomUUID().toString().replace("-", "").substring(0, 16);

        try {
            MDC.put(TRACE_ID, traceId);

            if (response instanceof HttpServletResponse httpResponse) {
                httpResponse.setHeader("X-Trace-Id", traceId);
            }

            if (request instanceof HttpServletRequest httpRequest) {
                log.debug("→ {} {}", httpRequest.getMethod(), httpRequest.getRequestURI());
            }

            chain.doFilter(request, response);
        } finally {
            MDC.remove(TRACE_ID);
        }
    }
}
