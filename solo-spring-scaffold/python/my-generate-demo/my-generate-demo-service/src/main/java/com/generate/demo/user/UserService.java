package com.generate.demo.user;

import com.generate.demo.common.api.PageResult;
import com.generate.demo.user.dto.UserCreateRequest;
import com.generate.demo.user.dto.UserResponse;
import com.generate.demo.user.dto.UserUpdateRequest;

public interface UserService {

    UserResponse getById(Long id);

    PageResult<UserResponse> page(int page, int size);

    Long create(UserCreateRequest request);

    void update(UserUpdateRequest request);

    void delete(Long id);
}
