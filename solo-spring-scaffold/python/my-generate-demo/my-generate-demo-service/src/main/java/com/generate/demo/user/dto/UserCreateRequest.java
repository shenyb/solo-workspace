package com.generate.demo.user.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
@Schema(description = "创建用户请求")
public class UserCreateRequest {

    @NotBlank(message = "{user.username.notblank}")
    @Schema(description = "用户名", example = "zhangsan")
    private String username;

    @Email(message = "{user.email.invalid}")
    @Schema(description = "邮箱", example = "zhangsan@example.com")
    private String email;
}
