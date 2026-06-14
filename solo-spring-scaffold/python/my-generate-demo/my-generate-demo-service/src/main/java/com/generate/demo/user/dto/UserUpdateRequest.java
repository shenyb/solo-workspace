package com.generate.demo.user.dto;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
@Schema(description = "更新用户请求")
public class UserUpdateRequest {

    @NotNull(message = "{user.id.notnull}")
    @Schema(description = "用户ID", example = "1")
    private Long id;

    @NotBlank(message = "{user.username.notblank}")
    @Schema(description = "用户名", example = "zhangsan")
    private String username;

    @Email(message = "{user.email.invalid}")
    @Schema(description = "邮箱", example = "zhangsan@example.com")
    private String email;
}
