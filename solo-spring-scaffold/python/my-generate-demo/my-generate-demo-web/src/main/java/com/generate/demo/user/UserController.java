package com.generate.demo.user;

import com.generate.demo.common.api.PageResult;
import com.generate.demo.common.api.Result;
import com.generate.demo.user.dto.UserCreateRequest;
import com.generate.demo.user.dto.UserResponse;
import com.generate.demo.user.dto.UserUpdateRequest;
import com.generate.demo.user.UserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import jakarta.validation.constraints.Min;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@Tag(name = "用户管理", description = "用户 CRUD 接口")
@RestController
@RequestMapping(value = "/api/users", produces = "application/json")
@RequiredArgsConstructor
public class UserController {

    private final UserService userService;

    @Operation(summary = "查询用户详情", description = "根据ID查询用户信息")
    @GetMapping("/{id}")
    public Result<UserResponse> getById(@PathVariable Long id) {
        return Result.success(userService.getById(id));
    }

    @Operation(summary = "分页查询用户列表")
    @GetMapping
    public Result<PageResult<UserResponse>> page(
            @RequestParam(defaultValue = "1") @Min(1) int page,
            @RequestParam(defaultValue = "10") @Min(1) int size) {
        return Result.success(userService.page(page, size));
    }

    @Operation(summary = "创建用户")
    @PostMapping
    public Result<Long> create(@Valid @RequestBody UserCreateRequest request) {
        return Result.success(userService.create(request));
    }

    @Operation(summary = "更新用户")
    @PutMapping
    public Result<Void> update(@Valid @RequestBody UserUpdateRequest request) {
        userService.update(request);
        return Result.success();
    }

    @Operation(summary = "删除用户")
    @DeleteMapping("/{id}")
    public Result<Void> delete(@PathVariable Long id) {
        userService.delete(id);
        return Result.success();
    }
}
