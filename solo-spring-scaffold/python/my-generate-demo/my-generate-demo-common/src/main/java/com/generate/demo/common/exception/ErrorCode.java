package com.generate.demo.common.exception;

import lombok.Getter;

@Getter
public enum ErrorCode {

    // ─── 通用 ───
    BAD_REQUEST(400, "请求参数错误"),
    UNAUTHORIZED(401, "未授权"),
    FORBIDDEN(403, "禁止访问"),
    NOT_FOUND(404, "资源不存在"),
    METHOD_NOT_ALLOWED(405, "请求方法不支持"),
    TOO_MANY_REQUESTS(429, "请求过于频繁"),

    // ─── 业务 ───
    BIZ_ERROR(1000, "业务异常"),
    DATA_EXISTS(1001, "数据已存在"),
    DATA_NOT_FOUND(1002, "%s不存在"),
    OPERATION_FAILED(1003, "操作失败"),

    // ─── 系统 ───
    INTERNAL_ERROR(5000, "系统内部错误"),
    DB_ERROR(5001, "数据库异常"),
    SERVICE_UNAVAILABLE(5002, "服务不可用"),
    THIRD_PARTY_ERROR(5003, "第三方服务异常");

    private final int code;
    private final String message;

    ErrorCode(int code, String message) {
        this.code = code;
        this.message = message;
    }
}
