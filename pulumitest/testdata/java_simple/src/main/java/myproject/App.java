package myproject;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.random.Integer;
import com.pulumi.random.IntegerArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Create a random integer between 1 and 100
            var randomNumber = new Integer("randomNumber", IntegerArgs.builder()
                .min(1)
                .max(100)
                .build());

            // Export the random number
            ctx.export("randomNumber", randomNumber.result());
        });
    }
}
