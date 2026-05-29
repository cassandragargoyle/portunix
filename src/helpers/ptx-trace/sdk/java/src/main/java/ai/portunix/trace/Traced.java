/*
 * This file is part of CassandraGargoyle Community Project
 * Licensed under the MIT License - see LICENSE file for details
 */
package ai.portunix.trace;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Annotation for tracing methods.
 *
 * <p>
 * Note: This annotation requires AspectJ or similar AOP framework to work.
 * Without AOP, use the explicit tracing API instead.</p>
 *
 * <p>
 * Example with AspectJ:</p>
 * <pre>{@code
 * @Traced(operation = "validate_email", tags = {"validation"})
 * public boolean validateEmail(String email) {
 *     return email.contains("@");
 * }
 * }</pre>
 *
 * <p>
 * The aspect would intercept calls and create trace operations automatically.</p>
 */
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface Traced
{
	/**
	 * Operation name. Defaults to the method name if empty.
	 */
	String operation() default "";

	/**
	 * Operation type.
	 */
	String type() default "transform";

	/**
	 * Tags to add to the operation.
	 */
	String[] tags() default
	{
	};
}
